package command

const loginPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>UU跑腿授权登录</title>
  <style>
    :root { color-scheme: light; }
    body {
      margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
      background: #fff4ec;
    }
    .card {
      width: 92%; max-width: 420px; background: #fff; border-radius: 16px;
      box-shadow: 0 10px 40px rgba(255, 90, 0, .12); padding: 28px 24px 24px;
    }
    h1 { margin: 0 0 6px; font-size: 22px; color: #222; }
    p.sub { margin: 0 0 20px; color: #888; font-size: 14px; }
    label { display: block; font-size: 13px; color: #555; margin: 12px 0 6px; }
    input {
      width: 100%; box-sizing: border-box; height: 42px; border: 1px solid #eadfd6;
      border-radius: 10px; padding: 0 12px; font-size: 15px; outline: none;
    }
    input:focus { border-color: #ff5a00; }
    .row { display: flex; gap: 8px; align-items: center; }
    .row input { flex: 1; }
    button {
      height: 42px; border: 0; border-radius: 10px; background: #ff5a00; color: #fff;
      font-size: 15px; cursor: pointer; padding: 0 14px;
    }
    button:disabled { opacity: .55; cursor: not-allowed; }
    button.ghost { background: #fff; color: #ff5a00; border: 1px solid #ff5a00; white-space: nowrap; }
    .msg { min-height: 20px; margin-top: 12px; font-size: 13px; color: #c0392b; }
    .msg.ok { color: #1e9e4a; }
    .captcha { display: none; margin-top: 8px; }
    .captcha img { height: 42px; border-radius: 8px; background: #f6f6f6; }
  </style>
</head>
<body>
  <div class="card">
    <h1>UU跑腿授权</h1>
    <p class="sub">请使用手机号完成短信验证，授权后即可在 WorkBuddy 中使用跑腿配送与帮帮服务。</p>
    <label>手机号</label>
    <input id="mobile" type="tel" maxlength="11" placeholder="11 位手机号" />
    <div class="captcha" id="captchaBox">
      <label>图片验证码</label>
      <div class="row">
        <input id="imageCode" maxlength="8" placeholder="请输入图中字符" />
        <img id="captchaImg" alt="验证码" />
      </div>
    </div>
    <label>短信验证码</label>
    <div class="row">
      <input id="smsCode" maxlength="8" placeholder="6 位验证码" />
      <button class="ghost" id="sendBtn" type="button">发送验证码</button>
    </div>
    <div style="margin-top:18px">
      <button id="loginBtn" type="button" style="width:100%">完成授权</button>
    </div>
    <div class="msg" id="msg"></div>
  </div>
  <script>
    const $ = (id) => document.getElementById(id);
    const msg = (text, ok) => { const el = $("msg"); el.textContent = text || ""; el.className = "msg" + (ok ? " ok" : ""); };
    const post = async (url, body) => {
      const res = await fetch(url, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || "请求失败");
      return data;
    };
    $("sendBtn").onclick = async () => {
      try {
        msg("");
        const data = await post("/api/sms", { mobile: $("mobile").value.trim(), imageCode: $("imageCode").value.trim() });
        if (data.needImage) {
          $("captchaBox").style.display = "block";
          let src = data.imageData || "";
          if (src && src.indexOf("data:") !== 0) src = "data:image/png;base64," + src;
          $("captchaImg").src = src;
          msg(data.message || "请输入图片验证码");
          return;
        }
        msg("验证码已发送，请查收短信", true);
      } catch (e) { msg(e.message); }
    };
    $("loginBtn").onclick = async () => {
      try {
        msg("");
        const data = await post("/api/verify", { mobile: $("mobile").value.trim(), smsCode: $("smsCode").value.trim() });
        msg("授权成功，可以关闭此页面返回 WorkBuddy", true);
        $("loginBtn").disabled = true;
        $("sendBtn").disabled = true;
      } catch (e) { msg(e.message); }
    };
  </script>
</body>
</html>
`
