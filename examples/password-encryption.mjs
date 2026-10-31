// 浏览器模块示例，无第三方依赖。Web Crypto 需要 HTTPS 或 localhost 安全上下文。
// fields 使用接口原有 snake_case 字段；返回值可直接 JSON.stringify 后提交。
export async function encryptPasswords(publicInfo, purpose, fields) {
  if (!globalThis.crypto?.subtle) throw new Error("当前环境不支持 Web Crypto，请使用 HTTPS 或 localhost");
  if (!["login", "register"].includes(purpose)) throw new Error("无效的认证用途");
  if (publicInfo.algorithm !== "RSA-OAEP-256") throw new Error("不支持的密码加密算法");
  const encoded = publicInfo.public_key.replace(/-----[^-]+-----/g, "").replace(/\s/g, "");
  const der = Uint8Array.from(atob(encoded), char => char.charCodeAt(0));
  const key = await crypto.subtle.importKey("spki", der, { name: "RSA-OAEP", hash: "SHA-256" }, false, ["encrypt"]);
  const nonce = Array.from(crypto.getRandomValues(new Uint8Array(16)), byte => byte.toString(16).padStart(2, "0")).join("");
  const encryption = { key_id: publicInfo.key_id, timestamp: Math.floor(Date.now() / 1000), nonce };
  const encrypt = async (value, field) => {
    if (typeof value !== "string") throw new Error("密码必须是字符串");
    const plaintext = new TextEncoder().encode(value);
    if ([...value].length < 8 || plaintext.length > 72) throw new Error("密码至少8个字符，且不超过72字节");
    const label = ["go-gin", "password", "v1", purpose, encryption.key_id, String(encryption.timestamp), nonce, fields.username, field].join(":");
    const cipher = await crypto.subtle.encrypt({ name: "RSA-OAEP", label: new TextEncoder().encode(label) }, key, plaintext);
    return btoa(String.fromCharCode(...new Uint8Array(cipher)));
  };
  const body = { ...fields, encryption, password: await encrypt(fields.password, "password") };
  delete body.password_confirmation;
  if (purpose === "register") body.password_confirmation = await encrypt(fields.password_confirmation, "password_confirmation");
  return body;
}
