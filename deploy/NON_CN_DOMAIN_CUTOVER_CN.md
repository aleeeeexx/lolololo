# 换非 .cn 对外域名（暂缓执行）

本文是 DubDub AI 中转站生产环境的可执行操作说明。先不要购买域名、不要改 DNS、不要改 Cloudflare、不要改生产配置。等用户量大了，或报告「打不开」的人明显变多时，再按下面步骤做。

当前对外地址 `https://api.dubdub-ai.cn` 必须先保留解析和服务。切过去之后，旧地址仍然要能打开，直到确认不再需要。

## 1. 背景和何时执行

### 现状

- 生产站：`https://api.dubdub-ai.cn`
- 解析路径：`api.dubdub-ai.cn` 经 Cloudflare **橙云（代理开启）** 指向 Hostinger 马来西亚 VPS `189.31.39.190`
- Cloudflare 账号：`alexforwork914131@gmail.com`
- 现有 zone：`dubdub-ai.cn`，zone id `2e1d9f4bf92d772858c1b97b86309cd8`（这是 `.cn` 的 zone；新域名要在同一账号下**新建** zone，不能把 `dubdub-ai.com` 的记录写进这个 `.cn` zone）

### 为什么以后可能要换

- 部分国内线路直连 Cloudflare Anycast（如 `104.21.*`、`172.67.*`）会被重置，浏览器显示连接中断。
- 很多代理会把 `*.cn` 当国内站走 `DIRECT`，请求仍打到被重置的线路上；不开 VPN 也不一定能修好。
- 换一个不会被当成国内站直连的非 `.cn` 域名，可以改善上述「按域名走直连」的情况，但不能保证全国直连（见第 7 节）。

### 何时执行

同时满足再动手：

1. 用户量已经明显增加，或多次收到「不开 VPN 打不开」的反馈；
2. 已再次确认推荐域名仍可注册（见第 2 节），并已决定购买哪一个；
3. 有完整维护窗口，能按第 6 节验证、按第 8 节回滚。

用户量还不多时：**不要执行本文后续步骤。**

## 2. 推荐域名（先核对再买）

查价与可注册性核对时间： **2026-11-25 14:00（Asia/Shanghai，CST）**。下单前必须再查一次 WHOIS/RDAP 和注册商搜索页，避免名字已被注册或变成溢价词。

### 选用顺序

| 优先级 | 域名 | 2026-11-25 状态 | 说明 |
|--------|------|-----------------|------|
| 第一推荐 | `dubdub-ai.com` | RDAP 未注册 | 与现有 `dubdub-ai.cn` 完全对应，作对外主域名 |
| 次选 | `getdubdub.com` | RDAP 未注册 | 第一推荐被占或溢价时用 |
| 次选 | `usedubdub.com` | RDAP 未注册 | 同上 |

下文把最终买到的域名写成 `<新域>`，对外主机名建议固定为 `api.<新域>`，例如 `api.dubdub-ai.com`。

### 不要用这些当对外主域名

- **不要用 `dubdub-ai.xyz` 当对外主域名。** 即便腾讯云能注册、首年也便宜，部分省份/运营商的反诈网关会拦 `.xyz` 这类廉价后缀，常见情况是同一台源站上 `.com` 能打开、`.xyz` 打不开。这和「方便国内直连」的目标相反。最多只能防御性占名，不能替代 `.com`。
- 不要用 `.cn` / `.com.cn` / `.net.cn` 再做一套对外主域，解决不了代理把 `.cn` 直连的问题。

### 已占用，不要买、不要找高价转卖

| 域名 | 公开登记情况（2026-11-25） |
|------|------------------------------|
| `dubdub.ai` | Namecheap，2021-04-25 注册，2027-04-25 到期；Observe.AI 收购的 AI 配音公司在用 |
| `dubdubai.com` | Cloudflare Registrar，2021-11-13 注册，2031-11-13 到期；站点为分类信息/市集，不是空域 |
| `godubdub.com` | GoDaddy，2026-01-21 注册；状态含 delete/renew/transfer/update prohibited |
| `mydubdub.com` | Namecheap，2024-02-14 注册，2027-02-14 到期 |

## 3. 价格参考（不要编造腾讯云 .com 官价）

以下为 2026-11-25 14:00（Asia/Shanghai）核对到的公开报价，实际以各注册商结算页为准。普通 `.com` 标准价适用于上面三个推荐名字（当时未见溢价标记）。

- **腾讯云（优先，用户已在用，国内微信支付/支付宝最方便）**  
  实时价以控制台为准：<https://buy.cloud.tencent.com/domain>  
  官网公告 <https://cloud.tencent.com/announce/detail/2514> 说明 2026-11-01 起上调 `.com` 新注/续费/转入，但**未公布调整后金额**。本文不填写腾讯云 `.com` 官价，下单时看控制台。
- **阿里云**  
  媒体转述口径（阿里云 2026-10-22 宣布、2026-11-01 起）：首年 90 元，续费 99 元/年。必须以官网结算页为准：<https://domain.aliyun.com/> 、<https://wanwang.aliyun.com/help/price.html>
- **Porkbun**  
  标准 `.com` 注册/续费/转入均为 **11.81 美元/年**（官方计价接口）。查询页示例：<https://porkbun.com/checkout/search?q=dubdub-ai.com>
- **Namecheap**  
  `.com` 首年促销 **12.08 美元**，续费 **19.98 美元/年**，结账时另加 **ICANN 0.20 美元**。商品页：<https://www.namecheap.com/domains/registration/gtld/com/>

建议：在腾讯云注册并继续用该账号管理域名，支付最省事。若腾讯云缺货或价格明显高于上述公开价，再对比阿里云或 Porkbun，但仍把 NS 指到下面的 Cloudflare 账号。

## 4. 执行前检查（禁止事项）

开始前确认：

- [ ] 再次确认 `<新域>` 可注册，且不是溢价词
- [ ] 能登录腾讯云域名控制台，以及 Cloudflare 账号 `alexforwork914131@gmail.com`
- [ ] 能登录源站，并已找到**现网** Caddy（或同等反代）配置，而不是仓库里的示例 `deploy/Caddyfile`
- [ ] 已备份现网 Caddy 配置，以及管理后台「站点设置」里当前的「API 端点地址」
- [ ] 准备好从国内不同线路（至少一条不用代理的线路）做第 6 节验证

执行全程遵守：

- **不要**改 `JWT_SECRET`、`TOTP_ENCRYPTION_KEY` 或其它密钥
- **不要**覆盖生产 `deploy/docker-compose.local.yml` 或 `deploy/.env`（需要改某个已有公钥地址时，只改那一项，不要整文件替换）
- **不要**无故重启 Postgres / Redis
- **不要**把 Cloudflare 改成灰云（仅 DNS）作为第一步，这会暴露源站 IP `189.31.39.190`
- **不要**删除或暂停 `api.dubdub-ai.cn` 的解析、证书或 Caddy 站点
- **不要**在未验证新主机名可用前，就把对外 API 地址改掉

## 5. 执行步骤

下面按顺序做。每一步完成后再做下一步。

### 步骤 A. 在腾讯云注册推荐域名

1. 打开 <https://buy.cloud.tencent.com/domain>，搜索第一推荐 `dubdub-ai.com`。
2. 若已不可注册或是溢价词，改搜次选 `getdubdub.com`，再不行则 `usedubdub.com`。不要改买 `dubdub-ai.xyz` 当主域。
3. 确认是普通标准价（不是溢价/保留字）后注册。建议一次注册多年，避免只盯首年低价。
4. 在腾讯云完成实名等注册商要求，直到域名状态为可管理。先不要改解析，等步骤 B 拿到 Cloudflare NS。

### 步骤 B. 在现有 Cloudflare 账号为新域新建 zone，并把 NS 指过去

`dubdub-ai.cn` 的 zone id `2e1d9f4bf92d772858c1b97b86309cd8` 只能管 `.cn`。新域必须另建 zone。

1. 登录 Cloudflare，账号 `alexforwork914131@gmail.com`。
2. 对照现有 zone `dubdub-ai.cn`，记下 SSL/TLS 模式（应为对橙云源站安全的模式，通常是 Full 或 Full (strict)）、始终 HTTPS、最低 TLS 版本、以及 `api.dubdub-ai.cn` 的记录类型和代理状态，后续要对齐，不要凭记忆改严或改松。
3. 添加站点 / 添加 zone，域名为刚买的 `<新域>`（例如 `dubdub-ai.com`）。这会生成一个**新的** zone id，与 `.cn` 的 id 不同，属正常。
4. Cloudflare 会给出该 zone 的两个 NS（形如 `*.ns.cloudflare.com`）。复制这对 NS。
5. 到腾讯云域名控制台，把 `<新域>` 的 DNS 服务器改成这对 Cloudflare NS。不要继续用腾讯云 DNS 再写一份 A 记录，以免和橙云冲突。
6. 等待 NS 生效。在确认 Cloudflare zone 状态变为 Active 之前，不要做对外切流。可用：

   ```bash
   dig +short NS <新域>
   ```

   结果应是 Cloudflare 刚给出的那对 NS，而不是腾讯云原 NS。

### 步骤 C. 为 `api.<新域>` 添加与 `api.dubdub-ai.cn` 对应的记录（橙云）

在**新域名的 zone**里操作，不要改 `.cn` zone 里已有记录。

1. 打开 `dubdub-ai.cn` zone，核对 `api` 主机记录：当前应为指向 `189.31.39.190` 的 A 记录，代理状态为橙云。若现网还有 AAAA 或其它同名记录，一并记下；没有就不要新造 AAAA。
2. 在 `<新域>` zone 中新增：
   - 名称：`api`
   - 类型：与现网 `api.dubdub-ai.cn` 相同（至少一条 A，目标 `189.31.39.190`）
   - 代理状态：**开启（橙云）**
3. 将新 zone 的 SSL/TLS、始终 HTTPS 等安全项对齐 `dubdub-ai.cn`。不要改成灰云。
4. 等解析出现 Cloudflare 代理地址（不是源站 IP）：

   ```bash
   dig +short A api.<新域>
   dig +short AAAA api.<新域>
   ```

   `dig +short A api.<新域>` 不应直接返回 `189.31.39.190`。若返回源站 IP，说明还在灰云或 NS 未生效，先停下排查，不要继续切对外地址。

### 步骤 D. 如需为新主机名增加 Caddy/证书站点

仓库里的 `deploy/Caddyfile` 只是示例（主机名为 `api.sub2api.com`），**不要把它覆盖到生产**。只改源站上正在使用的 Caddy 配置。

1. 在源站找到现网 Caddy 配置（常见为宿主机 Caddyfile，或单独的 Caddy 容器配置），确认已有 `api.dubdub-ai.cn` 站点块，以及它如何申请或加载证书（Let’s Encrypt、Cloudflare Origin Certificate 等）。
2. 备份该文件。
3. 为 `api.<新域>` 增加站点，反向代理、超时、压缩、转发头等必须与现网 `api.dubdub-ai.cn` 一致。优先把两个主机名写在同一个站点块，例如：

   ```caddy
   api.dubdub-ai.cn, api.<新域> {
       # 保持与现网 api.dubdub-ai.cn 完全相同的 tls、reverse_proxy 和其它指令
   }
   ```

   若证书是 Cloudflare Origin Certificate：在 Cloudflare 为 `api.<新域>`（或 `*.<新域>` 与 `<新域>`）重新签发或扩展证书，再装到源站，避免只把主机名写进 Caddy 却仍用仅含 `.cn` 的旧证书。
4. 校验配置后热加载，例如 `caddy validate` 再 `caddy reload`。只重载 Caddy，**不要**重启 Postgres、Redis，也不要 `docker compose down` 整套栈。
5. 在源站本机或内网先确认反代仍指向应用的 `/health`（现网 Caddy 示例使用 `health_uri /health`）。

### 步骤 E. 验证新主机名后再改对外展示地址

先保留 `api.dubdub-ai.cn`，确认新主机名可用，再改展示地址。

1. 按第 6 节检查 `https://api.<新域>/health` 和页面登录。新旧两个主机名都要过。
2. 管理员登录后台 → **站点设置** → **API 端点地址**（字段 `api_base_url`）。该值会出现在「使用密钥」「导入到 CC Switch」和用户可复制的 API 地址中；留空则用当前页面地址。
3. 把「API 端点地址」改为 `https://api.<新域>`（不要擅自加 `/v1` 或 `/api/v1`，除非现网原来就带有意后缀）。
4. 建议在「自定义端点」中保留 `https://api.dubdub-ai.cn`，方便仍使用旧地址的用户复制。
5. 若生产已经配置了 `MCP_PUBLIC_URL` 等带主机名的公钥地址，只改这一项到 `https://api.<新域>/api/v1/mcp`（路径规则以现网为准），不要整份替换 `deploy/.env`。若本来就是空的，不要为了换域名去新开 MCP。
6. 若 CORS `allowed_origins` 不是空列表，把 `https://api.<新域>` 加进去，同时保留 `https://api.dubdub-ai.cn`。改完如需加载新环境变量，只重启应用容器/进程，不要重启 Postgres/Redis。
7. 保存后台设置后，用普通用户登录，确认密钥页复制到的地址已是新域，且旧域名页面仍可打开。

## 6. 验证清单

新旧主机名都要查：

```bash
# 健康检查（现网 Caddy 使用 /health，期望 HTTP 200）
curl -sS -D- -o /tmp/health-cn.txt -w '%{http_code}\n' https://api.dubdub-ai.cn/health
curl -sS -D- -o /tmp/health-new.txt -w '%{http_code}\n' https://api.<新域>/health

# 证书主机名必须匹配
curl -svI https://api.<新域>/health 2>&1 | grep -E 'subject:|issuer:|SSL certificate verify|HTTP/'

# 解析必须走 Cloudflare 代理，不能露出源站 IP
dig +short A api.<新域>
```

再在浏览器中确认：

- [ ] `https://api.<新域>` 能打开登录页，证书无报错
- [ ] 能登录，侧边栏和「直接用」等页面正常
- [ ] 密钥页复制到的 API 地址是 `https://api.<新域>`
- [ ] `https://api.dubdub-ai.cn` 仍然可打开、可登录
- [ ] Cloudflare 两边记录都是橙云，`dig` 看不到 `189.31.39.190`

若新主机名健康检查失败，不要改「API 端点地址」，先按第 8 节保持对外仍为旧地址并排查。

## 7. 残余风险（不要写成全国直连）

换非 `.cn` 域名可以改善：

- 代理把 `*.cn` 当成国内站走 `DIRECT`；
- 部分按域名或证书做的限制。

**不能保证全国直连。** 若某些运营商是按 Cloudflare Anycast IP（如 `104.21.*`、`172.67.*`）重置，新域名仍然挂在橙云上，这些线路不开 VPN 仍可能打不开。不要对用户承诺「买了 `dubdub-ai.com` 就一定能打开」。

长期若要进一步提高国内可达性，再单独评估 ICP 备案，并走国内或香港加速/CDN。不要把 Cloudflare 灰云当作第一优化手段。

## 8. 回滚：对外地址切回 `api.dubdub-ai.cn`

新域出现问题、或暂时不再把新域当对外入口时：

1. 管理后台 → **站点设置** → 把 **API 端点地址** 改回 `https://api.dubdub-ai.cn` 并保存。
2. 若改过 `MCP_PUBLIC_URL` 或 CORS 白名单，把对外主机名改回 `https://api.dubdub-ai.cn`（CORS 可暂时同时保留新域来源）。只改对应项，不覆盖整份 `.env` 或 compose。
3. 不要删除 `api.<新域>` 的 Cloudflare 记录，也不要停掉 Caddy 里的新主机名，除非已经确认不再使用。保留它们便于再次切过去，也避免误伤旧证书。
4. 用第 6 节清单验证：`https://api.dubdub-ai.cn/health`、登录页、密钥页复制地址均已回到旧域名。
5. 不要为了回滚而改 JWT/TOTP，不要重启 Postgres/Redis，不要把橙云改成灰云。

`api.dubdub-ai.cn` 的解析和站点应一直保留，作为回滚目标和兼容入口。
