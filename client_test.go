package gowss

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	go func() {
		NewHub()
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			ServeWs(w, r)
		})
		err := http.ListenAndServe(":8090", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}

func newClient() *websocket.Conn {
	s := httptest.NewServer(http.HandlerFunc(ServeWs))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")
	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatalf("%v", err)
		return nil
	}
	// defer ws.Close()
	return ws
}

func TestClient(t *testing.T) {

	Convey("hello testing", t, func() {
		ws := newClient()
		defer ws.Close()
		// Send message to server, read response and check to see if it's what we expect.

		err := ws.WriteMessage(websocket.TextMessage, []byte("hello"))
		So(err, ShouldBeNil)

		So(len(MainHub.clients), ShouldEqual, 1)
		ws.Close()
		time.Sleep(time.Second * time.Duration(2))
		So(len(MainHub.clients), ShouldEqual, 0)
	})

	Convey("客户端注册属性添加", t, func() {
		ws := newClient()
		defer ws.Close()

		subM := `{"attrs":["kline.m1.demo", "latest.price.demo"]}`
		t.Log(subM)
		if err := ws.WriteMessage(websocket.TextMessage, []byte(subM)); err != nil {
			t.Fatalf("%v", err)
		}

		time.Sleep(time.Second * time.Duration(1))
		So(len(MainHub.clients), ShouldEqual, 1)
		for c, _ := range MainHub.clients {
			So(c.attrs, ShouldContainKey, "kline.m1.demo")
			So(c.attrs, ShouldContainKey, "latest.price.demo")
		}
	})

	Convey("给拥有订阅属性的客户端发送消息", t, func() {
		ws := newClient()
		defer ws.Close()

		subM := `{"attrs":["kline.m1.demo", "latest.price.demo"]}`
		t.Log(subM)
		if err := ws.WriteMessage(websocket.TextMessage, []byte(subM)); err != nil {
			t.Fatalf("%v", err)
		}

		send := MsgBody{
			To:   "kline.m1.demo",
			Body: []byte("hello"),
		}

		MainHub.Broadcast <- send
		_, recv, _ := ws.ReadMessage()

		time.Sleep(time.Second * time.Duration(1))

		So(len(MainHub.clients), ShouldEqual, 1)
		t.Logf("%s", recv)
		So(string(recv), ShouldEqualJSON, `{"type":"kline.m1.demo","body":"hello"}`)

		for c, _ := range MainHub.clients {
			So(c.lastSendMsgHash["kline.m1.demo"], ShouldEqual, "5d41402abc4b2a76b9719d911017c592")
		}
	})

	Convey("同一类型的消息重复发送去重", t, func() {
		// ws := newClient()
		// defer ws.Close()

		// subM := `{"attrs":["kline.m1.demo", "latest.price.demo"]}`
		// t.Log(subM)
		// if err := ws.WriteMessage(websocket.TextMessage, []byte(subM)); err != nil {
		// 	t.Fatalf("%v", err)
		// }

		// send := MsgBody{
		// 	To:   "kline.m1.demo",
		// 	Body: []byte("hello"),
		// }

		// MainHub.Broadcast <- send
		// MainHub.Broadcast <- send

		// _, recv, _ := ws.ReadMessage()

		// time.Sleep(time.Second * time.Duration(1))

		// So(len(MainHub.clients), ShouldEqual, 1)
		// So(string(recv), ShouldEqual, "hello")

		// for c, _ := range MainHub.clients {
		// 	So(c.lastSendMsgHash["kline.m1.demo"], ShouldEqual, "5d41402abc4b2a76b9719d911017c592")
		// }
	})
}
