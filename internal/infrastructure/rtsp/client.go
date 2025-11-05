//go:build cgo

package rtsp

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"ip-camera-viewer/internal/domain/model"
	videodecoder "ip-camera-viewer/internal/infrastructure/video"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtph264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/pion/rtp"
)

type Client struct {
	rtspClient  *gortsplib.Client
	config      *model.StreamConfig
	isConnected bool
	isRunning   bool
	mutex       sync.RWMutex
	cancelFunc  context.CancelFunc
}

func NewClient(config *model.StreamConfig) *Client {
	return &Client{
		rtspClient:  &gortsplib.Client{},
		config:      config,
		isConnected: false,
		isRunning:   false,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isConnected {
		return nil
	}

	u, err := base.ParseURL(c.config.RTSPURI)
	if err != nil {
		return fmt.Errorf("ошибка парсинга URL: %v", err)
	}

	c.rtspClient = &gortsplib.Client{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	c.rtspClient.Scheme = u.Scheme
	c.rtspClient.Host = u.Host

	err = c.rtspClient.Start()
	if err != nil {
		return fmt.Errorf("ошибка запуска клиента: %v", err)
	}

	c.isConnected = true
	return nil
}

func (c *Client) StartStreaming(ctx context.Context, frameChannel chan<- *model.FrameData) error {
	c.mutex.Lock()
	if c.isRunning {
		c.mutex.Unlock()
		return fmt.Errorf("поток уже запущен")
	}

	streamCtx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	c.isRunning = true
	c.mutex.Unlock()

	defer func() {
		c.mutex.Lock()
		c.isRunning = false
		c.cancelFunc = nil
		c.mutex.Unlock()
	}()

	if !c.isConnected {
		return fmt.Errorf("не подключен к RTSP потоку")
	}

	u, err := base.ParseURL(c.config.RTSPURI)
	if err != nil {
		return fmt.Errorf("ошибка парсинга URL: %v", err)
	}

	desc, _, err := c.rtspClient.Describe(u)
	if err != nil {
		return fmt.Errorf("ошибка описания потока: %v", err)
	}

	var forma *format.H264
	medi := desc.FindFormat(&forma)
	if medi == nil {
		return fmt.Errorf("H264 контент не найден")
	}

	rtpDec, err := forma.CreateDecoder()
	if err != nil {
		return fmt.Errorf("ошибка создания декодера RTP: %v", err)
	}

	h264Dec := &videodecoder.H264Decoder{}
	err = h264Dec.Initialize()
	if err != nil {
		return fmt.Errorf("ошибка инициализации декодера H264: %v", err)
	}
	defer func() {
		if h264Dec != nil {
			h264Dec.Close()
		}
	}()

	if forma.SPS != nil {
		h264Dec.Decode([][]byte{forma.SPS})
	}
	if forma.PPS != nil {
		h264Dec.Decode([][]byte{forma.PPS})
	}

	_, err = c.rtspClient.Setup(desc.BaseURL, medi, 0, 0)
	if err != nil {
		return fmt.Errorf("ошибка настройки транспорта: %v", err)
	}

	var firstRandomAccess bool
	var packetMutex sync.Mutex

	c.rtspClient.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		select {
		case <-streamCtx.Done():
			return
		default:
		}

		packetMutex.Lock()
		defer packetMutex.Unlock()

		_, ok := c.rtspClient.PacketPTS(medi, pkt)
		if !ok {
			return
		}

		au, err := rtpDec.Decode(pkt)
		if err != nil {
			if err != rtph264.ErrNonStartingPacketAndNoPrevious && err != rtph264.ErrMorePacketsNeeded {
				log.Printf("Ошибка декодировки РТП пакетов: %v", err)
			}
			return
		}

		if !firstRandomAccess {
			if !h264.IsRandomAccess(au) {
				return
			}
			firstRandomAccess = true
		}

		img, err := h264Dec.Decode(au)
		if err != nil || img == nil {
			return
		}

		safeImage := c.createSafeImageCopy(img)
		if safeImage == nil {
			return
		}

		frame := &model.FrameData{
			Image:     safeImage,
			Timestamp: time.Now(),
		}

		select {
		case frameChannel <- frame:
		case <-streamCtx.Done():
			return
		default:
		}
	})

	_, err = c.rtspClient.Play(nil)
	if err != nil {
		return fmt.Errorf("ошибка запуска воспроизведения: %v", err)
	}

	log.Printf("RTSP поток запущен: %s", c.config.RTSPURI)

	select {
	case <-streamCtx.Done():
		log.Printf("RTSP поток остановлен по контексту: %s", c.config.RTSPURI)
	default:
	}
	err = c.rtspClient.Wait()
	if err != nil {
		log.Printf("Ошибка RTSP клиента: %v", err)
	}

	return nil
}

func (c *Client) createSafeImageCopy(src image.Image) image.Image {
	if src == nil {
		return nil
	}

	bounds := src.Bounds()
	if bounds.Empty() {
		return nil
	}

	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	return dst
}

func (c *Client) StopStreaming() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cancelFunc != nil {
		c.cancelFunc()
		c.cancelFunc = nil
	}

	c.isRunning = false
}

func (c *Client) Close() error {
	c.StopStreaming()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.rtspClient != nil {
		c.rtspClient.Close()
	}

	c.isConnected = false
	return nil
}
