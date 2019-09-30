package hub

import (
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"sync"
)

type Hub struct {
	metrics  map[string]*prometheus.Desc
	registry map[string]Channel
	mutex    sync.Mutex
}

func New() *Hub {
	metrics := map[string]*prometheus.Desc{
		"channel": prometheus.NewDesc("client_cnt", "total active channels in pool", []string{"channel", "channel_type"}, nil),
		"hub":     prometheus.NewDesc("channel_cnt", "total active channels in pool", []string{}, nil),
		//"channel": prometheus.NewDesc("channel", "total active channels in pool", []string{"count"}, nil),
	}
	return &Hub{registry: make(map[string]Channel), metrics: metrics}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	channel := query.Get("channel")
	chType := query.Get("type")
	clientCode := query.Get("id")
	if chType == "" {
		chType = "default"
	}
	if len(clientCode) == 0 {
		clientCode = uuid.New().String()
	}
	ch, ok := h.registry[channel]
	var err error
	if !ok {
		//从factory获取channel
		ch, err = Factory().GetChannel(chType, WithChannelName(channel), WithParams(query))
		if err != nil {
			return
		}
		go ch.Run(h)
		h.Register(ch)
	}

	//接收websocket链接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//注册客户端连接
	client := &Client{Code: clientCode, channel: ch, conn: conn, send: make(chan []byte, 256)}
	client.channel.Register() <- client

	//启动读写loop
	go client.WLoop()
	go client.RLoop()

}

func (h *Hub) Register(channel Channel) {
	h.mutex.Lock()
	h.registry[channel.Code()] = channel
	h.mutex.Unlock()
}

func (h *Hub) UnRegister(channel Channel) {
	h.mutex.Lock()
	delete(h.registry, channel.Code())
	h.mutex.Unlock()
}

//prometheus metrics抓取
func (h *Hub) Describe(ch chan<- *prometheus.Desc) {
	for _, val := range h.metrics {
		ch <- val
	}
}

func (h *Hub) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(h.metrics["hub"], prometheus.GaugeValue, float64(len(h.registry)))
	for _, val := range h.registry {
		ch <- prometheus.MustNewConstMetric(h.metrics["channel"], prometheus.GaugeValue, val.TotalClients(), val.Code(), val.Type())
	}
}
