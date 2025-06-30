# Makefile for building Lua dynamic libraries

# Detect OS
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

# Define file extensions for different OS
ifeq ($(OS),linux)
    LIB_PREFIX := lib
    LIB_EXT := so
    LIB_FLAGS = -shared -fPIC
else ifeq ($(OS),darwin)
    LIB_PREFIX := lib
    LIB_EXT := dylib
    LIB_FLAGS = -shared -fPIC
else
    LIB_PREFIX := 
    LIB_EXT := dll
    OS := windows
    LIB_FLAGS = -shared
endif

# Directories
LUA54_SRC = pkgs/lua54
LUA54_LIB_DIR = lua54/.lua/lib
LUA54_LIB = $(LUA54_LIB_DIR)/$(LIB_PREFIX)lua54.$(LIB_EXT)

# Compiler settings
CC = gcc
CFLAGS = -O2 -Wall -Wextra $(LIB_FLAGS)

# Platform-specific compiler flags
ifeq ($(OS),linux)
    CFLAGS += -DLUA_USE_LINUX
endif
ifeq ($(OS),darwin)
    CFLAGS += -DLUA_USE_MACOSX
endif
ifeq ($(OS),windows)
    CFLAGS += -DLUA_BUILD_AS_DLL
endif

# Lua 5.4 source files (excluding lua.c and luac.c which are for standalone programs)
LUA54_CORE_OBJS = \
	$(LUA54_SRC)/lapi.o $(LUA54_SRC)/lauxlib.o $(LUA54_SRC)/lbaselib.o \
	$(LUA54_SRC)/lcode.o $(LUA54_SRC)/lcorolib.o $(LUA54_SRC)/lctype.o \
	$(LUA54_SRC)/ldblib.o $(LUA54_SRC)/ldebug.o $(LUA54_SRC)/ldo.o \
	$(LUA54_SRC)/ldump.o $(LUA54_SRC)/lfunc.o $(LUA54_SRC)/lgc.o \
	$(LUA54_SRC)/linit.o $(LUA54_SRC)/liolib.o $(LUA54_SRC)/llex.o \
	$(LUA54_SRC)/lmathlib.o $(LUA54_SRC)/lmem.o $(LUA54_SRC)/loadlib.o \
	$(LUA54_SRC)/lobject.o $(LUA54_SRC)/lopcodes.o $(LUA54_SRC)/loslib.o \
	$(LUA54_SRC)/lparser.o $(LUA54_SRC)/lstate.o $(LUA54_SRC)/lstring.o \
	$(LUA54_SRC)/lstrlib.o $(LUA54_SRC)/ltable.o $(LUA54_SRC)/ltablib.o \
	$(LUA54_SRC)/ltm.o $(LUA54_SRC)/lundump.o $(LUA54_SRC)/lutf8lib.o \
	$(LUA54_SRC)/lvm.o $(LUA54_SRC)/lzio.o

.PHONY: all clean tests-54 setup-dirs

all: tests-54

# Create necessary directories
setup-dirs:
	@mkdir -p $(LUA54_LIB_DIR)

# Build Lua 5.4 dynamic library
tests-54: setup-dirs $(LUA54_LIB)

$(LUA54_LIB): $(LUA54_CORE_OBJS)
	$(CC) $(CFLAGS) -o $@ $^

# Compile Lua 5.4 object files
$(LUA54_SRC)/%.o: $(LUA54_SRC)/%.c
	$(CC) $(CFLAGS) -c $< -o $@

# Clean build artifacts
clean:
	rm -f $(LUA54_SRC)/*.o
	rm -rf .lua

# Help target
help:
	@echo "Available targets:"
	@echo "  tests-54    - Build Lua 5.4 dynamic library for testing"
	@echo "  clean       - Remove build artifacts and .lua directory"
	@echo "  help        - Show this help message"

