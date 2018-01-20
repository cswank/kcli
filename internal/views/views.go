package views

import (
	"fmt"
	"log"
	"os"

	"github.com/cswank/kcli/internal/kafka"
	ui "github.com/jroimartin/gocui"
)

func NewGui(topic string, partition, offset int) error {
	g, err := ui.NewGui(ui.Output256)
	if err != nil {
		return fmt.Errorf("could not create gui: %s", err)
	}

	w, h := g.Size()
	opts := getOpts(h-2, topic, partition, offset)
	s, err := newScreen(g, w, h, opts...)
	if err != nil {
		g.Close()
		kafka.Close()
		log.Fatalf("error: %s", err)
	}

	g.SetManagerFunc(s.GetLayout(g, w, h))
	g.Cursor = true
	g.InputEsc = true

	if err := s.Keybindings(g); err != nil {
		return err
	}

	if err != nil {
		log.Fatal(err)
	}

	var closed bool
	defer func() {
		if !closed {
			g.Close()
		}
		kafka.Close()
	}()

	if err := g.MainLoop(); err != nil {
		if err != ui.ErrQuit {
			log.SetOutput(os.Stderr)
			log.Println(err)
			return err
		}
	}

	closed = true
	g.Close()
	if s.After != nil {
		s.After()
	}

	return nil
}

func getOpts(height int, topic string, partition, offset int) []func(*stack) error {
	var out []func(*stack) error
	if topic != "" {
		out = append(out, enterTopic(height, topic))
		if partition != -1 {
			out = append(out, enterPartitionOrOffset(height, partition))
			if offset != -1 {
				out = append(out, enterPartitionOrOffset(height, offset))
			}
		}
	}
	return out
}
