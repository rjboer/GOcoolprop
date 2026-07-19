package reference

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type Case struct {
	Output string  `json:"output"`
	Input1 string  `json:"input1"`
	Input2 string  `json:"input2"`
	Fluid  string  `json:"fluid"`
	Value1 float64 `json:"value1"`
	Value2 float64 `json:"value2"`
}
type Request struct {
	RequestID string `json:"request_id"`
	Fluid     string `json:"fluid"`
	Cases     []Case `json:"cases"`
}
type Item struct {
	Value float64 `json:"value"`
	Error string  `json:"error"`
	Phase string  `json:"phase"`
}
type Response struct {
	RequestID string `json:"request_id"`
	Results   []Item `json:"results"`
	Error     string `json:"error"`
}
type Worker struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     *bufio.Reader
	Startup map[string]any
	mu      sync.Mutex
}

func Start(ctx context.Context, python, script string) (*Worker, error) {
	cmd := exec.CommandContext(ctx, python, script)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	r := &Worker{cmd: cmd, in: in, out: bufio.NewReader(out), Startup: map[string]any{}}
	line, err := r.out.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(line, &r.Startup); err != nil {
		_ = r.in.Close()
		_ = r.cmd.Wait()
		return nil, fmt.Errorf("decode reference startup: %w", err)
	}
	if msg, ok := r.Startup["startup_error"].(string); ok && msg != "" {
		_ = r.in.Close()
		_ = r.cmd.Wait()
		return nil, fmt.Errorf("reference startup error: %s", msg)
	}
	return r, nil
}
func (w *Worker) Call(req Request) (Response, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	b, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	if _, err = w.in.Write(append(b, '\n')); err != nil {
		return Response{}, err
	}
	line, err := w.out.ReadBytes('\n')
	if err != nil {
		return Response{}, err
	}
	var r Response
	if err = json.Unmarshal(line, &r); err != nil {
		return Response{}, fmt.Errorf("decode reference response: %w", err)
	}
	return r, nil
}
func (w *Worker) Close() error { _ = w.in.Close(); return w.cmd.Wait() }
