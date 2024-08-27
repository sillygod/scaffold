package handlers

import (
	"exampleproj/cache"
	"exampleproj/events"
	"exampleproj/internal/app"
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type subscriber struct {
	Controller *events.AppController
	rdb        *redis.Client
}

func (s subscriber) PingRequestOperationReceived(ctx context.Context, ping events.PingMessage) error {
	// Publish the pong message, with the callback function to modify it
	// Note: it will indefinitely wait to publish as context has no timeout
	err := s.Controller.ReplyToPingRequestOperation(ctx, ping, func(pong *events.PongMessage) {
		// Reply a pong message
		res := "pong"
		pong.Payload.Event = &res
	})

	// Error management
	if err != nil {
		panic(err)
	}

	return nil
}

func (s subscriber) PricefeedRequestOperationReceived(ctx context.Context, pricefeed events.PricefeedRequestMessage) error {
	err := s.Controller.ReplyToPricefeedRequestOperation(ctx, pricefeed, func(pricefeed *events.PricefeedMessage) {
		ctx := context.Background()

		event := "pricefeed"
		pricefeed.Payload.Event = &event
		priceLists := []events.ItemFromPriceListPropertyFromPricefeedMessagePayload{}

		for _, feedId := range app.FeedIds {

			data, err := cache.GetAllStreamEntries(ctx, s.rdb, "pyth_history_price_feed_"+feedId)

			if err != nil {
				panic(err)
			}

			feedData := events.ItemFromPriceListPropertyFromPricefeedMessagePayload{
				FeedId: &feedId,
			}

			for _, entry := range data {
				price, err := strconv.ParseFloat(entry.Values["price"].(string), 64)
				if err != nil {
					panic(err)
				}
				feedData.Prices = append(feedData.Prices, price)

				ts, err := strconv.ParseInt(entry.Values["ts"].(string), 10, 64)
				if err != nil {
					panic(err)
				}
				feedData.Timestamps = append(feedData.Timestamps, ts)
			}

			priceLists = append(priceLists, feedData)
		}

		pricefeed.Payload.PriceList = priceLists

	})

	return err
}

// define a websocket handler that matched the interface of routers.Handler
type WebsocketHandler struct {
	hub *app.Hub
	rdb *redis.Client
}

func NewWebsocketHandler(lc fx.Lifecycle, rdb *redis.Client) *WebsocketHandler {

	hub := app.NewHub()

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go hub.Run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return &WebsocketHandler{
		hub: hub,
		rdb: rdb,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (ws *WebsocketHandler) handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		client := app.NewWSClient(ws.hub, conn, 512)

		ctrl, err := events.NewAppController(client)
		if err != nil {
			panic(err)
		}

		sb := subscriber{
			Controller: ctrl,
			rdb:        ws.rdb,
		}

		client.BindAppController(ctrl)
		ctrl.SubscribeToAllChannels(client.Context(), sb)
	}
}

func (ws *WebsocketHandler) RegisterRoute(r *chi.Mux) {
	r.Get("/ws", ws.handle())
}

var _ Handler = (*WebsocketHandler)(nil)
