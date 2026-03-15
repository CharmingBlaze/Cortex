# GTK4 GUI Library Makefile
#
# Build: make
# Clean: make clean
# Install: make install

CC = gcc
CFLAGS = -Wall -Wextra -O2 -fPIC
GTK_CFLAGS = $(shell pkg-config --cflags gtk4)
GTK_LIBS = $(shell pkg-config --libs gtk4)

# Source files
SRCS = gui_core.c gui_widgets.c gui_containers.c gui_dialogs.c
OBJS = $(SRCS:.c=.o)

# Output
LIB = libgui_gtk4.a
SHARED = libgui_gtk4.so

# Default target
all: $(LIB) $(SHARED)

# Static library
$(LIB): $(OBJS)
	ar rcs $@ $^

# Shared library
$(SHARED): $(OBJS)
	$(CC) -shared -o $@ $^ $(GTK_LIBS)

# Compile
%.o: %.c gui_gtk4_internal.h
	$(CC) $(CFLAGS) $(GTK_CFLAGS) -c -o $@ $<

# Clean
clean:
	rm -f $(OBJS) $(LIB) $(SHARED)

# Install to system
install: $(LIB) $(SHARED)
	install -d $(DESTDIR)/usr/local/lib
	install -d $(DESTDIR)/usr/local/include
	install -m 644 $(LIB) $(DESTDIR)/usr/local/lib/
	install -m 755 $(SHARED) $(DESTDIR)/usr/local/lib/
	install -m 644 ../../runtime/gui_runtime.h $(DESTDIR)/usr/local/include/

# Test build
test: $(LIB)
	$(CC) -o test_gui test_gui.c -L. -lgui_gtk4 $(GTK_LIBS)

.PHONY: all clean install test
