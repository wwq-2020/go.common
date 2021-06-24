package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

var (
	setupOnce sync.Once
	globalApp *App
	// GraceExitHook GraceExitHook
	GraceExitHook = func() {
	}
)

// App 应用
type App struct {
	ctx context.Context
	sync.Mutex
	children      []*App
	cancel        func()
	wg            sync.WaitGroup
	shutdownHooks []func()
}

// New 创建应用
func New() *App {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		ctx:    ctx,
		cancel: cancel,
	}
	return app
}

// Go 开启应用goroutine
func (app *App) Go(f func(context.Context)) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		f(app.ctx)
	}()
}

// GoForever GoForever
func (app *App) GoForever(f func(context.Context)) {
	app.wg.Add(1)
	wrappedF := func(ctx context.Context) {
		defer func() {
			if e := recover(); e != nil {
				stack := stack.Callers(stack.StdFilter)
				var err error
				switch t := e.(type) {
				case error:
					err = t
				default:
					err = fmt.Errorf("%#v", t)
				}
				log.WithField("stack", stack).
					ErrorContext(ctx, err)
			}
		}()
		f(ctx)
	}
	go func() {
		for {
			wrappedF(app.ctx)
			select {
			case <-app.ctx.Done():
				app.wg.Done()
				return
			default:
			}
		}
	}()
}

// GoAsync 开启应用goroutine
func (app *App) GoAsync(f func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		f()
	}()
}

// GoAsyncForever GoAsyncForever
func (app *App) GoAsyncForever(f func()) {
	app.wg.Add(1)
	wrappedF := func() {
		defer func() {
			if e := recover(); e != nil {
				stack := stack.Callers(stack.StdFilter)
				var err error
				switch t := e.(type) {
				case error:
					err = t
				default:
					err = fmt.Errorf("%#v", t)
				}
				log.WithField("stack", stack).
					Error(err)
			}
		}()
		f()
	}
	go func() {
		for {
			wrappedF()
			select {
			case <-app.ctx.Done():
				app.wg.Done()
				return
			default:
			}
		}
	}()
}

// GoWithContext 开启应用goroutine,携带context
func (app *App) GoWithContext(ctx context.Context, f func(context.Context)) {
	app.wg.Add(2)
	wrappedCtx, cancel := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	go func() {
		defer func() {
			app.wg.Done()
		}()
		select {
		case <-ctx.Done():
			cancel()
		case <-app.ctx.Done():
			cancel()
		case <-doneCh:
		}
	}()
	go func() {
		defer func() {
			close(doneCh)
			app.wg.Done()
		}()
		f(wrappedCtx)
	}()
}

// Close 结束这个应用
func (app *App) Close() {
	app.Lock()
	defer app.Unlock()
	for _, child := range app.children {
		child.Close()
		child.Wait()
	}
	app.cancel()
	for _, shutdownHook := range app.shutdownHooks {
		shutdownHook()
	}
}

// Wait 等待应用goroutine全部退出
func (app *App) Wait() {
	app.wg.Wait()
}

// AddChild 添加子应用
func (app *App) AddChild(child *App) {
	app.Lock()
	app.children = append(app.children, child)
	app.Unlock()
}

// AddShutdownHook 添加退出hook
func (app *App) AddShutdownHook(hook func()) {
	app.Lock()
	app.shutdownHooks = append(app.shutdownHooks, hook)
	app.Unlock()
}

// CatchExitSignal 捕获退出信号
func CatchExitSignal(callback func()) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	signal := <-signalCh
	go func() {
		signal := <-signalCh
		log.WithField("signal", signal).
			Info("force exit")
		os.Exit(1)
	}()
	GraceExitHook = func() {
		log.WithField("signal", signal).
			Info("grace exit")
	}
	callback()
}

// AddShutdownHook 添加退出hook
func AddShutdownHook(hook func()) {
	setupOnce.Do(setup)
	globalApp.AddShutdownHook(hook)
}

// setup 初始化全局app
func setup() {
	globalApp = New()
	go CatchExitSignal(globalApp.Close)
}

// Go 开启应用goroutine
func Go(f func(context.Context)) {
	setupOnce.Do(setup)
	globalApp.Go(f)
}

// GoForever GoForever
func GoForever(f func(context.Context)) {
	setupOnce.Do(setup)
	globalApp.GoForever(f)
}

// GoAsyncForever GoAsyncForever
func GoAsyncForever(f func()) {
	setupOnce.Do(setup)
	globalApp.GoAsyncForever(f)
}

// GoAsync GoAsync
func GoAsync(f func()) {
	setupOnce.Do(setup)
	globalApp.GoAsync(f)
}

// Wait 等待应用goroutine全部退出
func Wait() {
	setupOnce.Do(setup)
	select {
	case <-Done():
		Close()
		return
	default:
	}
	globalApp.Wait()
	GraceExitHook()
}

// Close 结束应用
func Close() {
	setupOnce.Do(setup)
	globalApp.Close()
}

// Done Done
func Done() <-chan struct{} {
	setupOnce.Do(setup)
	return globalApp.ctx.Done()
}

// Context Context
func Context() context.Context {
	setupOnce.Do(setup)
	return globalApp.ctx
}

// AddChild 添加子应用
func AddChild(child *App) {
	setupOnce.Do(setup)
	globalApp.AddChild(child)
}
