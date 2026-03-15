// Package ide provides IDE-like features for Cortex
//
// Integrations:
// - github.com/fsnotify/fsnotify: File watching for hot reload
// - github.com/spf13/afero: Virtual filesystem for sandboxed projects
// - github.com/spf13/viper: Configuration management

package ide

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// ============================================
// Reactive Data Binding (native implementation)
// ============================================

// StringBinding is a simple reactive string
type StringBinding struct {
	mu        sync.RWMutex
	value     string
	listeners []func(string)
}

// NewStringBinding creates a new reactive string
func NewStringBinding() *StringBinding {
	return &StringBinding{
		listeners: make([]func(string), 0),
	}
}

// Get returns the current value
func (b *StringBinding) Get() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.value
}

// Set updates the value and notifies listeners
func (b *StringBinding) Set(v string) {
	b.mu.Lock()
	b.value = v
	listeners := make([]func(string), len(b.listeners))
	copy(listeners, b.listeners)
	b.mu.Unlock()

	for _, l := range listeners {
		l(v)
	}
}

// Listen adds a change listener
func (b *StringBinding) Listen(fn func(string)) {
	b.mu.Lock()
	b.listeners = append(b.listeners, fn)
	b.mu.Unlock()
}

// StringListBinding is a simple reactive string list
type StringListBinding struct {
	mu    sync.RWMutex
	value []string
}

// NewStringList creates a new reactive string list
func NewStringList() *StringListBinding {
	return &StringListBinding{value: make([]string, 0)}
}

// Get returns the current value
func (l *StringListBinding) Get() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.value
}

// Set updates the value
func (l *StringListBinding) Set(v []string) {
	l.mu.Lock()
	l.value = v
	l.mu.Unlock()
}

// BoolBinding is a simple reactive bool
type BoolBinding struct {
	mu    sync.RWMutex
	value bool
}

// NewBool creates a new reactive bool
func NewBool() *BoolBinding {
	return &BoolBinding{}
}

// Get returns the current value
func (b *BoolBinding) Get() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.value
}

// Set updates the value
func (b *BoolBinding) Set(v bool) {
	b.mu.Lock()
	b.value = v
	b.mu.Unlock()
}

// ============================================
// Output Buffer
// ============================================

// OutputBuffer is a reactive string buffer for compiler output/logs
type OutputBuffer struct {
	inner *StringBinding
	mu    sync.Mutex
	lines []string
}

// NewOutputBuffer creates a reactive output buffer
func NewOutputBuffer() *OutputBuffer {
	return &OutputBuffer{
		inner: NewStringBinding(),
		lines: make([]string, 0),
	}
}

// Append adds a line and triggers UI update
func (o *OutputBuffer) Append(line string) {
	o.mu.Lock()
	o.lines = append(o.lines, line)
	text := o.inner.Get()
	o.mu.Unlock()
	o.inner.Set(text + line + "\n")
}

// Clear resets the buffer
func (o *OutputBuffer) Clear() {
	o.mu.Lock()
	o.lines = nil
	o.mu.Unlock()
	o.inner.Set("")
}

// Get returns the current content
func (o *OutputBuffer) Get() string {
	return o.inner.Get()
}

// Listen adds a change listener
func (o *OutputBuffer) Listen(fn func(string)) {
	o.inner.Listen(fn)
}

// ============================================
// Filesystem Watching (fsnotify/fsnotify)
// ============================================

// FileWatcher monitors files for changes
type FileWatcher struct {
	watcher   *fsnotify.Watcher
	callbacks map[string]func()
	mu        sync.RWMutex
}

// NewFileWatcher creates a file watcher
func NewFileWatcher() (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{
		watcher:   w,
		callbacks: make(map[string]func()),
	}, nil
}

// Watch adds a path to monitor with callback
func (fw *FileWatcher) Watch(path string, callback func()) error {
	fw.mu.Lock()
	fw.callbacks[path] = callback
	fw.mu.Unlock()
	return fw.watcher.Add(path)
}

// Start begins watching for changes
func (fw *FileWatcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fw.mu.RLock()
					if cb, exists := fw.callbacks[event.Name]; exists {
						go cb()
					}
					fw.mu.RUnlock()
				}
			case _, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}

// Close stops the watcher
func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}

// ============================================
// Virtual Filesystem (spf13/afero)
// ============================================

// VirtualFS provides an in-memory filesystem for sandboxed projects
type VirtualFS struct {
	Fs afero.Fs
}

// NewVirtualFS creates an in-memory filesystem
func NewVirtualFS() *VirtualFS {
	return &VirtualFS{
		Fs: afero.NewMemMapFs(),
	}
}

// NewOsFS creates a real filesystem wrapper
func NewOsFS() *VirtualFS {
	return &VirtualFS{
		Fs: afero.NewOsFs(),
	}
}

// WriteFile writes content to a virtual file
func (vfs *VirtualFS) WriteFile(path, content string) error {
	return afero.WriteFile(vfs.Fs, path, []byte(content), 0644)
}

// ReadFile reads content from a virtual file
func (vfs *VirtualFS) ReadFile(path string) (string, error) {
	data, err := afero.ReadFile(vfs.Fs, path)
	return string(data), err
}

// Exists checks if a file exists
func (vfs *VirtualFS) Exists(path string) bool {
	exists, _ := afero.Exists(vfs.Fs, path)
	return exists
}

// ListDir lists files in a directory
func (vfs *VirtualFS) ListDir(path string) ([]string, error) {
	files, err := vfs.Fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer files.Close()
	names, err := files.Readdirnames(-1)
	return names, err
}

// ============================================
// Configuration (spf13/viper)
// ============================================

// Config manages Cortex IDE settings
type Config struct {
	v *viper.Viper
}

// NewConfig creates a configuration manager
func NewConfig() *Config {
	v := viper.New()
	v.SetConfigName("cortex")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.cortex")

	// Defaults
	v.SetDefault("editor.font_size", 14)
	v.SetDefault("editor.theme", "dark")
	v.SetDefault("editor.tab_size", 4)
	v.SetDefault("compiler.backend", "gcc")
	v.SetDefault("compiler.optimization", "-O2")
	v.SetDefault("build.auto_save", true)
	v.SetDefault("build.output_dir", "build")

	return &Config{v: v}
}

// Load reads configuration from file
func (c *Config) Load() error {
	return c.v.ReadInConfig()
}

// Save writes configuration to file
func (c *Config) Save() error {
	return c.v.WriteConfig()
}

// Get retrieves a config value
func (c *Config) Get(key string) interface{} {
	return c.v.Get(key)
}

// GetString retrieves a string config value
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt retrieves an int config value
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool retrieves a bool config value
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// Set stores a config value
func (c *Config) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

// ============================================
// Project Manager (combines all integrations)
// ============================================

// Project represents a Cortex project with IDE features
type Project struct {
	Name     string
	Path     string
	FS       *VirtualFS
	Watcher  *FileWatcher
	Output   *OutputBuffer
	Config   *Config
	Files    *StringListBinding
	Modified *BoolBinding
}

// NewProject creates a new project with all IDE integrations
func NewProject(name, path string) (*Project, error) {
	watcher, err := NewFileWatcher()
	if err != nil {
		return nil, err
	}

	p := &Project{
		Name:     name,
		Path:     path,
		FS:       NewVirtualFS(),
		Watcher:  watcher,
		Output:   NewOutputBuffer(),
		Config:   NewConfig(),
		Files:    NewStringList(),
		Modified: NewBool(),
	}

	// Auto-reload on file change
	watcher.Watch(path, func() {
		p.Modified.Set(true)
		p.Output.Append("File changed, reloading...")
	})
	watcher.Start()

	return p, nil
}

// AddFile adds a file to the project
func (p *Project) AddFile(name, content string) error {
	if err := p.FS.WriteFile(name, content); err != nil {
		return err
	}
	files := p.Files.Get()
	files = append(files, name)
	p.Files.Set(files)
	return nil
}

// Compile logs compiler output to the reactive buffer
func (p *Project) Compile() {
	p.Output.Clear()
	p.Output.Append("Compiling " + p.Name + "...")
	p.Output.Append("Backend: " + p.Config.GetString("compiler.backend"))
	p.Output.Append("Flags: " + p.Config.GetString("compiler.optimization"))
	// Simulate compilation
	time.Sleep(100 * time.Millisecond)
	p.Output.Append("Build complete!")
	p.Modified.Set(false)
}

// Close cleans up resources
func (p *Project) Close() error {
	return p.Watcher.Close()
}
