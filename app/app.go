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

// Done Done
func (app *App) Done() <-chan struct{} {
	return app.Context().Done()
}

// Context Context
func (app *App) Context() context.Context {
	return app.ctx
}

// Go 开启应用goroutine
func (app *App) Go(f func(context.Context)) {
	wrappedF := func() {
		f(app.Context())
	}
	app.GoAsync(wrappedF)
}

// GoForever GoForever
func (app *App) GoForever(f func(context.Context)) {
	app.wg.Add(1)
	wrappedF := func() {
		f(app.Context())
	}
	app.GoAsync(func() {
		defer app.wg.Done()
		for {
			wrappedF()
			select {
			case <-app.ctx.Done():
				return
			default:
			}
		}
	})
}

// withCancel withCancel
func (app *App) withCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	app.wg.Add(1)
	newCtx, cancel := context.WithCancel(ctx)
	app.GoAsync(func() {
		defer app.wg.Done()
		defer cancel()
		select {
		case <-newCtx.Done():
		case <-app.ctx.Done():
		}
	})
	return newCtx, cancel
}

// GoForeverContext GoForeverContext
func (app *App) GoForeverContext(ctx context.Context, f func(context.Context)) {
	newCtx, _ := app.withCancel(ctx)
	wrappedF := func() {
		f(newCtx)
	}
	app.GoAsyncForeverContext(ctx, wrappedF)
}

// GoAsync 开启应用goroutine
func (app *App) GoAsync(f func()) {
	app.wg.Add(1)
	go func() {
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
			app.wg.Done()
		}()
		f()
	}()
}

// GoAsyncForever GoAsyncForever
func (app *App) GoAsyncForever(f func()) {
	app.GoAsyncForeverContext(context.Background(), f)
}

// GoAsyncForeverContext GoAsyncForeverContext
func (app *App) GoAsyncForeverContext(ctx context.Context, f func()) {
	app.wg.Add(1)
	wrappedF := func() {
		f()
	}
	app.GoAsync(func() {
		defer app.wg.Done()
		for {
			wrappedF()
			select {
			case <-app.ctx.Done():
				return
			case <-ctx.Done():
				return
			default:
			}
		}
	})
}

// GoContext 开启应用goroutine,携带context
func (app *App) GoContext(ctx context.Context, f func(context.Context)) {
	newCtx, cancel := app.withCancel(ctx)
	wrappedF := func() {
		defer cancel()
		f(newCtx)
	}
	app.GoAsync(wrappedF)
}

// Close 结束这个应用
func (app *App) Close() {
	app.Lock()
	defer app.Unlock()
	for _, child := range app.children {
		child.Close()
		child.Wait()
	}
	for _, shutdownHook := range app.shutdownHooks {
		shutdownHook()
	}
	app.cancel()
	log.Info("grace exit")
}

// Wait 等待应用goroutine全部退出
func (app *App) Wait() {
	select {
	case <-app.Done():
		app.Close()
		app.wg.Wait()
		log.Info("grace exit")
		return
	default:
	}
	app.wg.Wait()
	log.Info("grace exit")
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
	<-signalCh
	go func() {
		signal := <-signalCh
		log.WithField("signal", signal).
			Info("force exit")
		os.Exit(1)
	}()
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

// GoForeverContext GoForeverContext
func GoForeverContext(ctx context.Context, f func(context.Context)) {
	setupOnce.Do(setup)
	globalApp.GoForeverContext(ctx, f)
}

// GoAsyncForever GoAsyncForever
func GoAsyncForever(f func()) {
	setupOnce.Do(setup)
	globalApp.GoAsyncForever(f)
}

// GoAsyncForeverContext GoAsyncForeverContext
func GoAsyncForeverContext(ctx context.Context, f func()) {
	setupOnce.Do(setup)
	globalApp.GoAsyncForeverContext(ctx, f)
}

// GoAsync GoAsync
func GoAsync(f func()) {
	setupOnce.Do(setup)
	globalApp.GoAsync(f)
}

// Wait 等待应用goroutine全部退出
func Wait() {
	setupOnce.Do(setup)
	globalApp.Wait()
}

// Close 结束应用
func Close() {
	setupOnce.Do(setup)
	globalApp.Close()
}

// Done Done
func Done() <-chan struct{} {
	setupOnce.Do(setup)
	return globalApp.Done()
}

// Context Context
func Context() context.Context {
	setupOnce.Do(setup)
	return globalApp.Context()
}

// AddChild 添加子应用
func AddChild(child *App) {
	setupOnce.Do(setup)
	globalApp.AddChild(child)
}
