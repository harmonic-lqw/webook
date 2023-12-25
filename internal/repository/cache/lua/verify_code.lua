local key = KEYS[1]

local cntKey = key..":cnt"
-- 用户输入的验证码
local expectedCode = ARGV[1]

local cnt = tonumber(redis.call("get", cntKey))

local code = redis.call("get", key)

-- 验证次数耗尽
if cnt == nil or cnt <= 0 then
	return -1
end

if code == expectedCode then
	-- 把次数标记为 -1， 认为验证码不可用
	redis.call("set", cntKey, -1)
	return 0
else
    redis.call("decr", cntKey)
    return -2
end
