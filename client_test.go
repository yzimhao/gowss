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

var _socket *Hub

func init() {
	go func() {
		_socket = NewHub()
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			_socket.ServeWs(w, r)
		})
		err := http.ListenAndServe(":8090", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}

func newClient() *websocket.Conn {
	s := httptest.NewServer(http.HandlerFunc(_socket.ServeWs))
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
		err := ws.WriteMessage(websocket.TextMessage, []byte("hello"))
		So(err, ShouldBeNil)

		time.Sleep(time.Second * time.Duration(1))
		So(len(_socket.clients), ShouldEqual, 1)
		ws.Close()
		time.Sleep(time.Second * time.Duration(2))
		So(len(_socket.clients), ShouldEqual, 0)
	})

	Convey("客户端注册属性添加", t, func() {
		ws := newClient()
		defer ws.Close()

		subM := `{"sub":["kline.m1.demo", "latest.price.demo"]}`
		t.Log(subM)
		if err := ws.WriteMessage(websocket.TextMessage, []byte(subM)); err != nil {
			t.Fatalf("%v", err)
		}

		time.Sleep(time.Second * time.Duration(1))
		So(len(_socket.clients), ShouldEqual, 1)
		for c, _ := range _socket.clients {
			So(c.attrs, ShouldContainKey, "kline.m1.demo")
			So(c.attrs, ShouldContainKey, "latest.price.demo")
		}
	})

	Convey("给拥有订阅属性的客户端发送消息", t, func() {
		ws := newClient()
		defer ws.Close()

		subM := `{"sub":["kline.m1.demo", "latest.price.demo"]}`
		t.Log(subM)
		if err := ws.WriteMessage(websocket.TextMessage, []byte(subM)); err != nil {
			t.Fatalf("%v", err)
		}

		send := MsgBody{
			To: "kline.m1.demo",
			Body: []string{
				"a", "b",
			},
		}

		_socket.Broadcast <- send
		_, recv, _ := ws.ReadMessage()

		time.Sleep(time.Second * time.Duration(1))

		So(len(_socket.clients), ShouldEqual, 1)
		t.Logf("%s", recv)
		So(string(recv), ShouldEqualJSON, `{"type":"kline.m1.demo","body":["a","b"]}`)

		for c, _ := range _socket.clients {
			So(c.lastSendMsgHash["kline.m1.demo"], ShouldEqual, "f2534fe3f8a3ffd8243077e8d354eb17")
		}
	})

	Convey("同一类型的消息重复发送去重", t, func() {

	})
}
