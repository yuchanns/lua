local lm = require("luamake")

local sources
do
	local function build_dir(version)
		local dir = "pkgs/lua" .. version
		return function(path, exclude)
			local source = dir .. "/" .. path
			if exclude then
				source = "!" .. source
			end
			return source
		end
	end
	sources = function(version, paths)
		local srcs = {}
		local dir = build_dir(version)
		for _, path in pairs(paths) do
			srcs[#srcs + 1] = dir(path.path, path.exclude)
		end
		return srcs
	end
end

local function lua_dll(version)
	local lua_version = "lua" .. version
	local bindir = lua_version .. "/.lua/lib"
	lm:dll(lua_version)({
		sources = sources(version, {
			{ path = "*.c" },
			{ path = "onelua.c", exclude = true },
			{ path = "lua.c", exclude = true },
			{ path = "luac.c", exclude = true },
			{ path = "ltests.c", exclude = true },
		}),
		includes = {
			"pkgs/" .. lua_version,
		},
		bindir = bindir,

		c = "c99",

		visibility = "default",
		links = {
			"m",
		},

		windows = {
			defines = {
				"LUA_BUILD_AS_DLL",
			},
		},

		macos = {
			defines = {
				"LUA_USE_MACOSX",
			},
		},

		linux = {
			defines = {
				"LUA_USE_LINUX",
			},
		},

		gcc = {
			flags = {
				"-fPIC",
			},
		},

		clang = {
			flags = {
				"-fPIC",
			},
		},
	})
	local output = bindir
		.. "/"
		.. (
			lm.os == "windows" and lua_version .. ".dll"
			or (lm.os == "macos" and "lib" .. lua_version .. ".dylib" or "lib" .. lua_version .. ".so")
		)
	if lm.os ~= "windows" then
		lm:copy("copy_" .. lua_version)({
			deps = { lua_version },
			inputs = { bindir .. "/" .. lua_version .. ".so" },
			outputs = { output },
		})
	end
end

for _, version in ipairs({ 54, 53 }) do
	lua_dll(version)
end
