package lua

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
)

// RunPrivateProtocolLuaScript 执行用户 Lua 脚本
// luaScript: 用户自定义 Lua 脚本，必须定义 parse(input) 函数
// input: 原始字节数组，Lua 脚本自己解析
// 返回: JSON 序列化后的 []byte 和 logs和 error
func RunPrivateProtocolLuaScript(luaScript string, input []byte) (map[string]interface{}, []string, error) {
	L := lua.NewState()
	defer L.Close()

	var logs []string

	//-------------------------------------------------
	// 1. 向 Lua 注册 log(msg) 函数
	//-------------------------------------------------
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		logs = append(logs, msg)
		return 0
	}))

	//-------------------------------------------------
	// 2. 加载 Lua 脚本
	//-------------------------------------------------
	if err := L.DoString(luaScript); err != nil {
		return nil, logs, fmt.Errorf("failed to load lua script: %w", err)
	}

	//-------------------------------------------------
	// 3. 将 []byte 转为 Lua 字符串
	//-------------------------------------------------
	luaInput := lua.LString(string(input))

	//-------------------------------------------------
	// 4. 调用 parse(input)
	//-------------------------------------------------
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("parse"),
		NRet:    2, // result, error
		Protect: true,
	}, luaInput); err != nil {
		return nil, logs, fmt.Errorf("lua parse call failed: %w", err)
	}

	luaResult := L.Get(-2)
	luaErr := L.Get(-1)
	L.Pop(2)
	// 检查 error
	if errStr, ok := luaErr.(lua.LString); ok && string(errStr) != "" {
		return nil, logs, fmt.Errorf("lua script error: %s", string(errStr))
	}

	// 将 Lua table 转为 Go map
	resultMap := make(map[string]interface{})
	if luaTable, ok := luaResult.(*lua.LTable); ok {
		luaTable.ForEach(func(k, v lua.LValue) {
			resultMap[k.String()] = v
		})
	} else {
		return nil, logs, fmt.Errorf("lua parse result is not a table")
	}
	return resultMap, logs, nil
}
