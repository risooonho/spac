package game

import (
	"github.com/20zinnm/spac/common/net"
	"github.com/20zinnm/spac/common/net/downstream"
	"log"
	"github.com/google/flatbuffers/go"
	"net/url"
	"github.com/gorilla/websocket"
	"fmt"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/client/fonts"
)

type ConnectingScene struct {
	next chan Scene
	dots int
	txt  *text.Text
	win  *pixelgl.Window
}

func (s *ConnectingScene) Update(dt float64) {
	select {
	case scene := <-s.next:
		fmt.Println("next scene (old:connecting)")
		CurrentScene = scene
	default:
		s.win.Clear(colornames.Black)
		s.txt.Clear()
		s.dots++
		l := "Connecting"
		if s.dots > 120 {
			s.dots = 0
		} else if s.dots > 100 {
			l += "."
		} else if s.dots > 80 {
			l += ".."
		} else if s.dots > 60 {
			l += "..."
		} else if s.dots > 40 {
			l += ".."
		} else if s.dots > 20 {
			l += "."
		}
		s.txt.Dot.X -= s.txt.BoundsOf(l).W() / 2
		fmt.Fprintf(s.txt, l)
		s.txt.Draw(s.win, pixel.IM.Moved(s.win.Bounds().Max.Scaled(.5)).Scaled(s.win.Bounds().Max.Scaled(.5), 2))
	}
}

func newConnecting(win *pixelgl.Window, host string) *ConnectingScene {
	scene := &ConnectingScene{
		next: make(chan Scene),
		txt:  text.New(pixel.ZV, fonts.Atlas),
		win:  win,
	}
	go func() {
		u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
		log.Printf("connecting to %s", u.String())
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		log.Print("connected")
		conn := net.Websocket(c)
		for {
			message, err := readMessage(conn)
			if err != nil {
				log.Fatalln(err)
			}
			if message.PacketType() != downstream.PacketServerSettings {
				log.Fatalln("received non-settings packet first; aborting")
			}
			packetTable := new(flatbuffers.Table)
			if !message.Packet(packetTable) {
				log.Fatalln("failed to decode settings packet")
			}
			settings := new(downstream.ServerSettings)
			settings.Init(packetTable.Bytes, packetTable.Pos)
			scene.next <- newMenu(win, conn)
			return
		}
	}()
	return scene
}
