package httpsender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/agent/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/FlutterDizaster/ya-metrics/pkg/validation"
	"github.com/FlutterDizaster/ya-metrics/pkg/workerpool"
	"github.com/go-resty/resty/v2"
)

// Настройки сервиса отправки метрик.
type Settings struct {
	Addr             string         // Адрес сервера агрегации метрик
	RetryCount       int            // Количество повторных попыток отправки метрик
	RetryInterval    time.Duration  // Интервал между повторными попытками
	RetryMaxWaitTime time.Duration  // Максимальное время ожидания между повторными попытками
	ReportInterval   time.Duration  // Интервал между отправками метрик
	HashKey          string         // Хеш ключ
	Buf              sender.Buffer  // Буфер метрик
	RateLimit        int            // Максимальное кол-во запросов в секунду
	RSAKey           *rsa.PublicKey // Сертификат TLS
}

// Sender - сервис отправки метрик.
// Должен быть создан через New.
type Sender struct {
	endpointAddr   string
	client         *resty.Client
	reportInterval time.Duration
	hashKey        string
	buf            sender.Buffer
	wpool          workerpool.WorkerPool
	rsaKey         *rsa.PublicKey
	hostAddr       string
}

// Фабрика создания экземпляра Sender.
func New(settings Settings) *Sender {
	slog.Debug("Creating sender")
	sender := &Sender{
		endpointAddr:   fmt.Sprintf("http://%s/updates/", settings.Addr),
		client:         resty.New(),
		reportInterval: settings.ReportInterval,
		hashKey:        settings.HashKey,
		buf:            settings.Buf,
		wpool:          *workerpool.New(settings.RateLimit),
		rsaKey:         settings.RSAKey,
	}
	sender.client.SetRetryCount(settings.RetryCount)
	sender.client.SetRetryWaitTime(settings.RetryInterval)
	sender.client.SetRetryMaxWaitTime(settings.RetryMaxWaitTime)

	hostAddr, err := getHostAddr()
	if err != nil {
		slog.Error("failed get host address", "error", err)
	}
	sender.hostAddr = hostAddr

	return sender
}

// getHostAddr - получение локального адреса хоста.
// В случае ошибки возвращается пустая строка.
// Предпологается, что локальная сеть включает в себя адреса между 192.168.0.0 и 192.168.255.255.
func getHostAddr() (string, error) {
	// Получение списка адресов хоста
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// Создание экземпляра IPNet для локальной сети
	localNet := &net.IPNet{
		IP:   net.IPv4(192, 168, 0, 0),
		Mask: net.IPv4Mask(255, 255, 0, 0),
	}

	// Поиск локального адреса хоста
	for _, addr := range addrs {
		// Проверка типа адреса и преобразование в IPNet
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// Проверка, что адрес является IPv4
			if ipnet.IP.To4() == nil {
				continue
			}

			// Проверка, что адрес принадлежит локальной сети
			if localNet.Contains(ipnet.IP) {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("host address not found")
}

// Start - запуск сервиса отправки метрик.
// Блокирует потов выполнения до завершения работы сервиса.
// Завершает работу сервиса при завершении контекста.
func (s *Sender) Start(ctx context.Context) error {
	slog.Debug("Sender", slog.String("status", "start"))
	ticker := time.NewTicker(s.reportInterval)
	slog.Info("Sender started", "report interval", s.reportInterval)

	// Первая отправка метрик
	s.send(ctx)

	// Старт основного цикла
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			lastCtx, lastCancleCtx := context.WithTimeout(context.Background(), 3*time.Second)
			defer lastCancleCtx()
			s.send(lastCtx)
			s.wpool.Close()
			slog.Debug("Sender", slog.String("status", "stop"))
			return nil
		case <-ticker.C:
			s.send(ctx)
		}
	}
}

func (s *Sender) send(ctx context.Context) {
	slog.Debug("Sender", slog.String("status", "sending..."))
	// ПОлучение метрик из буфера агента
	metrics, err := s.buf.Pull()
	if err != nil {
		slog.Error("Sender", "error", err)
		return
	}

	// Маршалинг метрик
	metricsBytes, err := view.Metrics(metrics).MarshalJSON()
	if err != nil {
		slog.Error("marshaling error", "error", err)
		return
	}

	// Формирование запроса
	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetContext(ctx)

	// Подсчет хеша при необходимости
	if s.hashKey != "" {
		hash := validation.CalculateHashSHA256(metricsBytes, []byte(s.hashKey))
		req.SetHeader("HashSHA256", hex.EncodeToString(hash))
	}

	// Сжатие метрики
	data, err := compressData(metricsBytes)
	if err != nil {
		slog.Error("compression error", "error", err)
		data = metricsBytes
	} else {
		req.SetHeader("Content-Encoding", "gzip")
	}

	// Шифрование при необходимости
	if s.rsaKey != nil {
		encryptedData, encErr := rsa.EncryptPKCS1v15(rand.Reader, s.rsaKey, data)
		if encErr != nil {
			slog.Error("encryption error", "error", err)
		} else {
			data = encryptedData
		}
	}

	// Установка тела запроса
	req.SetBody(data)

	// Установка X-Real-IP
	//
	// Устанавливается именно локальный адрес, так как даже если запрос
	// будет уходить во внешнюю сеть, то установкой правильного X-Real-IP должен
	// будет заниматься прокси сервер стоящий перед сервером обработчиком метрик.
	req.SetHeader("X-Real-IP", s.hostAddr)

	// Отправка запроса
	err = s.wpool.Do(func() {
		resp, errr := req.Post(s.endpointAddr)
		if errr != nil {
			slog.Info("Sender", "error", errr)
		} else {
			slog.Info(
				"Sender",
				slog.String("status", "sended"),
				slog.Int("response_code", resp.StatusCode()),
			)
		}
	})
	if err != nil {
		slog.Error("unexpected sender error", "error", err)
	}
}

func compressData(data []byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	gz, err := gzip.NewWriterLevel(buf, gzip.BestSpeed)
	if err != nil {
		slog.Error("failed init gzip writer", "error", err)
		return []byte{}, err
	}

	_, err = gz.Write(data)
	if err != nil {
		slog.Error("compress error", "error", err)
		return []byte{}, err
	}

	err = gz.Close()
	if err != nil {
		slog.Error("compress error", "error", err)
		return []byte{}, err
	}

	return buf.Bytes(), nil
}
