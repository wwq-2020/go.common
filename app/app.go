package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/syncx"
)

var (
	setupOnce sync.Once
	globalApp *App
	// GraceExitHook GraceExitHook
	GraceExitHook = func() {
	}
)

type Hook func(ctx context.Context) context.Context

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
func (app *App) Go(f func(context.Context), hooks ...Hook) {
	app.wg.Add(1)
	wrappedF := func() {
		defer app.wg.Done()
		ctx := app.Context()
		for _, hook := range hooks {
			ctx = hook(ctx)
		}
		f(ctx)
	}
	syncx.SafeGo(wrappedF)
}

// GoAsync 开启应用goroutine
func (app *App) GoAsync(f func()) {
	app.wg.Add(1)
	wrappedF := func() {
		defer app.wg.Done()
		f()
	}
	syncx.SafeGo(wrappedF)
}

// GoForever GoForever
func (app *App) GoForever(f func(context.Context), hooks ...Hook) {
	wrapped := func() {
		ctx := app.Context()
		for _, hook := range hooks {
			ctx = hook(ctx)
		}
		f(ctx)
	}
	onStart := func() { app.wg.Add(1) }
	onStop := func() { app.wg.Done() }
	syncx.SafeLoopGoex(app.Context(), wrapped, onStart, onStop)
}

// GoForever GoForever
func (app *App) GoAsyncForever(f func()) {
	onStart := func() { app.wg.Add(1) }
	onStop := func() { app.wg.Done() }
	syncx.SafeLoopGoex(app.Context(), f, onStart, onStop)
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
	defer app.Unlock()
	app.children = append(app.children, child)
}

// AddShutdownHook 添加退出hook
func (app *App) AddShutdownHook(hook func()) {
	app.Lock()
	defer app.Unlock()
	app.shutdownHooks = append(app.shutdownHooks, hook)
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
