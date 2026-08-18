# gcsim 内容扩充与项目复活执行清单

> 分析基线：`main` 分支 `45585d29d`（2026-07-05）。本文针对“新增几个角色、2 种反应、几套圣遗物、几十把武器”，名称和具体数量确定后，把占位符替换为实际条目即可。

## 0. 先说结论

- **先实现反应，再实现依赖这些反应的角色、圣遗物和武器。** 反应属于内核功能，不是普通内容包；它会穿过附着、触发顺序、伤害公式、事件、统计、日志和本地化。
- **不要做一个包含全部内容的巨型 PR。** 推荐 1 个“机制规格”PR、每种反应 1 个内核 PR、每个角色 1 个 PR、每套圣遗物 1 个 PR；武器按共享模板的小批次拆分。
- **先补测试和生成链路，再追求数量。** 当前 CI 只测试部分 Go 包，UI 流程只安装依赖；角色/武器/圣遗物几乎没有针对被动和数值的单元测试。
- **优先争取原项目的共同维护权，同时准备社区 fork。** MIT 许可证允许 fork，但域名、部署密钥、Cloudflare、数据库、Discord bot 和发布身份不会随代码一起转移。

当前执行文档：

- [内容矩阵](docs/content-matrix.md)
- [Community preview 固定数据快照](docs/community-data-snapshot.md)
- [不完整实现政策](docs/incomplete-implementation-policy.md)
- [星反应可执行规格门槛](docs/mechanics/stellar-reactions-spec.md)

## 1. 仓库结构和新增内容的真实链路

```text
外部游戏数据（ExcelBinOutput + TextMap）
                 |
                 v
 internal/**/config.yml ---> pipeline/cmd/generate
                 |                  |
                 |                  +--> *_gen.go / data_gen.textproto
                 |                  +--> pkg/core/keys、ICD、角色 imports
                 |                  +--> UI / DB JSON、文档页、本地化、资源映射
                 v
手写 Go 机制包 --init() 注册--> core registry
                 |
                 +--> pkg/simulation 的 blank imports 使 init() 真正执行
                 +--> pkg/shortcut 使 gcsim 配置文本能识别名字
                 +--> 文档中的 Frames/AoE/Hitlag/Issues 等人工数据
```

关键目录：

- 角色：`internal/characters/<key>/`
- 武器：`internal/weapons/<class>/<key>/`
- 圣遗物：`internal/artifacts/<key>/`
- 反应：`pkg/reactable/`
- 反应公共类型和公式：`pkg/core/info/`、`pkg/core/attacks/`、`pkg/core/combat/`
- 实际伤害/附着时序：`pkg/enemy/attack.go`、`pkg/enemy/damage.go`
- 注册表：`pkg/core/register.go`
- GCSL 名称解析：`pkg/shortcut/`
- 全局导入：`pkg/simulation/imports.go`、`pkg/simulation/imports_char_gen.go`
- Pipeline：`pipeline/cmd/generate/main.go`、`pipeline/pkg/`
- 前端数据：`ui/packages/ui/src/Data/`、`ui/packages/db/src/Data/`
- 文档：`ui/packages/docs/docs/reference/`、`ui/packages/docs/src/components/`

当前规模（用于估算改动冲突）：116 份角色 config（含旅行者形态）、52 个已实现圣遗物包、228 把武器。仓库已经有月感电、月绽放、月结晶，可作为“最近实现”的参考，但不要直接复制其中的特殊规则。

## 2. P0：开始写内容前必须完成

- [ ] 建立 `content-matrix`（GitHub Project 或表格），每个条目至少有：游戏 ID、内部 key、类型、依赖机制、实现人、研究人、复核人、证据链接、代码状态、测试状态、已知偏差。
- [ ] 为 2 种新反应分别写一页**可执行规格**，规格不清楚时不进入编码。
- [ ] 固定外部游戏数据仓库及其 commit/hash；不要让不同开发者用不同版本的 `ExcelBinOutput` 生成代码。
- [ ] 明确数据的获取、再分发和许可证边界。`pipeline/data/` 当前被忽略，仓库本身没有提供数据快照。
- [ ] 统一工具版本。`go.mod` 要求 Go 1.25，但测试 action 使用 `^1.21`，lint workflow 还写着 `^1.19.3`，应统一到 `go.mod`。
- [ ] 保存一组稳定的基准模拟：固定 config、seed、目标数量/位置、迭代数和期望统计。后续每个内核 PR 都跑差分。
- [ ] 先把现有 CI 绿色基线记录下来，再扩大测试范围；不要在新增内容 PR 中顺手修大量无关旧失败。
- [ ] 明确“不完整实现”的标记政策：哪些缺失会阻止合并，哪些可记录到 docs `Issues` 后合并。

建议的内容矩阵：

| 条目 | 类型 | 依赖 | 实现 | 机制复核 | 实机/数据复核 | 状态 |
|---|---|---|---|---|---|---|
| `<reaction-a>` | 反应 | 无 | A | B | C | spec |
| `<reaction-b>` | 反应 | `<reaction-a>` 或无 | B | A | C | backlog |
| `<character-a>` | 角色 | `<reaction-a>` | C | A | B | backlog |
| `<weapon-batch-1>` | 武器 | 通用模板 | D | B | C | backlog |

## 3. 两种新反应：先写规格，再改内核

### 3.1 每种反应必须回答的规格问题

- [ ] 哪两个（或多个）元素触发；正向/反向触发是否不同。
- [ ] 与已有 aura 的共存关系，消耗倍率、残留量、衰减、刷新和转化规则。
- [ ] 在 `Reactable.React` 中的优先级；一个攻击能否连续触发多个反应。
- [ ] 元素附着受哪个 ICD tag/group 控制，反应自身伤害是否另有 GCD/ICD。
- [ ] 属于增幅、剧变、激化、月反应式、直伤式还是全新公式。
- [ ] 等级、精通、反应加成、基础伤害加成、暴击、抗性、防御、伤害加成分别如何参与公式。
- [ ] 伤害归属谁：触发者、附着者、最后贡献者，还是多人贡献；换人/死亡后是否保留。
- [ ] 单目标、多目标、链式、范围、目标死亡、目标离场时如何处理。
- [ ] 是否生成 gadget/construct/状态，数量上限、替换规则、持续时间和命中盒是什么。
- [ ] 是否默认启用，还是像现有月反应一样由角色通过 `core.Flags.Custom` 开启。
- [ ] 触发事件发几次、携带什么参数；反应触发事件和反应伤害事件必须区分。
- [ ] 哪些旧角色、武器、圣遗物、共鸣会受它影响，哪些明确不应受影响。

规格中的每一个常量都应附数据来源、版本和置信度。无法验证的行为不要伪装成精确实现，使用明确的 TODO、issue 和可覆盖参数。

### 3.2 预计代码触点

- [ ] `pkg/core/info/reaction.go`：新增 `ReactionType`；需要公开能力时扩展 `Reactable` 子接口。
- [ ] `pkg/core/info/reactionmodifier.go`：只有新增 aura/modifier 才加 key；同时维护字符串和 element 映射。
- [ ] `pkg/core/event/event.go`：新增反应事件并放在 reaction delimiters 内；确认订阅参数保持 `target, *AttackEvent`。
- [ ] `pkg/core/attacks/attack.go`：只有产生独立伤害类型时增加 `AttackTag`，并确认 delimiter 分类函数。
- [ ] `pkg/reactable/<reaction>.go`：实现检测、消耗、状态、tick、伤害与归属。
- [ ] `pkg/reactable/reactable.go`：把反应插入正确的元素分支和优先级。
- [ ] `pkg/core/combat/reaction.go`：集中放可复用公式，避免在角色、反应和 `enemy` 中各复制一份倍率。
- [ ] `pkg/enemy/attack.go`：复核 `OnEnemyHit -> pre-damage mods -> React -> calc -> OnEnemyDamage/AttachOrRefill` 的时序。
- [ ] `pkg/enemy/damage.go`：仅当新伤害模型确实不同才增加专用分支。
- [ ] `pkg/stats/reaction/reaction.go`：事件到反应名映射。
- [ ] `ui/packages/localization/src/locales/*.json`：所有现有语言至少有稳定 fallback，不要只改 English/Chinese。
- [ ] 受影响的角色、圣遗物、武器、共鸣和条件系统。

注意：`event.Event`、`AttackTag`、`ReactionModKey` 都是 `iota` 枚举。插入位置会改变后续数字值；虽然它们主要是进程内使用，也要检查日志、缓存、WASM 或任何潜在序列化消费者。delimiter 的相对位置尤其不能错。

### 3.3 反应合并门槛

- [ ] 正向和反向触发测试。
- [ ] 0 aura、刚好耗尽、部分残留、多 aura 共存测试。
- [ ] 与每个竞争反应的优先级回归测试。
- [ ] 等级/EM/反应加成至少三组公式表驱动测试。
- [ ] ICD/GCD、首次 tick、后续 tick、结束帧测试。
- [ ] 贡献者归属、换人、重复触发、刷新和目标死亡测试。
- [ ] 1/2/3 个目标及不同距离测试。
- [ ] 启用与未启用测试。
- [ ] 事件恰好触发一次，统计能按角色正确记账。
- [ ] 现有反应测试全部通过，固定 seed 的基准模拟没有无关漂移。

测试优先参考 `pkg/reactable/*_test.go` 的轻量 target/core；需要完整队伍、装备或伤害管线时再放入 `internal/tests/`。

## 4. 新增角色

### 4.1 文件和注册清单

- [ ] 创建 `internal/characters/<key>/config.yml`。
- [ ] 填 `package_name`、`genshin_id`、`key`；旅行者/形态分支还要确认 `sub_id`、`key_var_name` 和目录约定。
- [ ] 在 `skill_data_mapping` 中把每个生成变量映射到游戏数据 param index；`-1` 只用于有意生成全零占位。
- [ ] 有动作参数时填 `action_param_keys`，否则非法参数不会在解析阶段被拒绝。
- [ ] 有专属 ICD 时填 `icd_groups` / `icd_tags`，名称必须全仓唯一。
- [ ] 有条件字段时填 `documentation.fields_data` 并实现 `Condition`。
- [ ] 编写 `<key>.go`，在 `init()` 中调用 `core.RegisterCharFunc(keys.<Key>, NewChar)`。
- [ ] `NewChar` 设置 EnergyMax、NormalHitNum、SkillCon/BurstCon、特殊阵营/机制，并把 wrapper 的 `Character` 指回实例。
- [ ] `Init()` 注册依赖完整队伍或完整基础属性的事件/被动。
- [ ] 按职责拆分 `attack.go`、`charge.go`/`aim.go`、`plunge.go`、`skill.go`、`burst.go`、`asc.go`、`cons.go`；不要做成一个数千行文件。
- [ ] 运行 character pipeline，提交生成的 `data_gen.textproto`、`<package>_gen.go`、角色 key、ICD、角色 imports、UI/DB 数据、资源映射和文档页；再用完整 pipeline 生成多语言名称 diff。
- [ ] 手工更新 `pkg/shortcut/characters.go`。
- [ ] 同时更新两个语法高亮副本：`ui/packages/components/src/Editor/mode-gcsim.js` 和 `ui/packages/ui/src/util/mode-gcsim.js`。当前两个文件已经有轻微顺序差异，后续应合并为单一数据源。
- [ ] 更新文档的 Actions、Frames、Hitlag、AoE、Issues、Params 等人工 JSON；pipeline 目前只自动写部分角色字段/页面/名称数据。

### 4.2 角色机制核对清单

- [ ] 普攻每段倍率、段数、hitmark、可取消帧、攻速调整、弹道/范围。
- [ ] 重击/瞄准/下落/特殊跳跃与 `NextQueueItemIsValid`。
- [ ] E/Q 的 snapshot frame、damage frame、动画结束、`CanQueueAfter`、CD 生效帧和耗能帧。
- [ ] `AttackTag`、`ICDTag`、`ICDGroup`、元素、durability、strike type、poise、hitlag。
- [ ] 粒子数量、概率、触发条件、粒子 ICD 和飞行延迟。
- [ ] buff 是动态还是快照、前台还是后台、是否跟随角色 hitlag、刷新/叠层/覆盖规则。
- [ ] task 应使用全局 `Core.Tasks.Add` 还是受角色 hitlag 影响的 `QueueCharTask`。
- [ ] 多目标命中、弹射、目标选择、目标死亡、生成物数量上限。
- [ ] A1/A4、C1-C6、C3/C5 天赋等级；联机专属效果可以不做，但必须文档化。
- [ ] Xingqiu/Yelan N0、Xianyun 下落、替代冲刺、特殊移动等跨角色钩子。
- [ ] C0 和 C6、天赋 1/10/13、前后台、单/多目标至少各有一条验证用例。

最近完整角色 PR 通常会改变约 30 个文件，其中很多是生成文件和文档数据；这不是异常，但应把“手写 diff”和“生成 diff”分成不同 commit，降低审核成本。

## 5. 新增几十把武器

### 5.1 先分类，避免复制几十份相似代码

- [ ] A 类：纯白板/无被动，复用 `internal/weapons/common.NewNoEffect`。
- [ ] B 类：已有通用族（西风、祭礼、黑岩、宗室等），先抽取或复用 shared helper，再按参数配置。
- [ ] C 类：独特被动，每把独立实现和测试。
- [ ] 不要为了“少写代码”把语义不同、时序不同的被动硬塞进一个万能 helper。

### 5.2 每把武器清单

- [ ] 创建 `internal/weapons/<class>/<key>/config.yml`：`package_name`、`genshin_id`、`key`，需要时加 `struct_name`/`skip_data_func`。schema 虽有 `shortcuts` 字段，但当前 generator 未消费它，别名仍须手工写入 `pkg/shortcut/weapons.go`。
- [ ] 编写被动并 `core.RegisterWeaponFunc(keys.<Weapon>, NewWeapon)`。
- [ ] 在 `pkg/core/keys/weapon.go` 同步维护 canonical string 和 Go const。两部分的顺序/数量必须一致，不能只加一边。
- [ ] 在 `pkg/shortcut/weapons.go` 加 canonical name 和无冲突别名。
- [ ] 在 `pkg/simulation/imports.go` 加 blank import。当前 weapon pipeline 不自动生成它。
- [ ] 运行 weapon pipeline，提交 `data_gen.textproto`、`<package>_gen.go`、UI JSON、资源映射和文档页；再用完整 pipeline 生成多语言名称数据。
- [ ] 手工补文档组件中的 Names，以及被动需要的 Params/Issues/AoE/Fields；pipeline 只生成引用这些组件的页面，不会推导被动行为。
- [ ] 验证武器类型与角色武器类型匹配时可解析、可构造、可运行。

### 5.3 武器测试矩阵

- [ ] R1 与 R5；线性精炼公式不要只测 R1。
- [ ] 触发攻击是否包含“触发这一击”本身。`OnEnemyHit` 在 pre-damage mods 之前，但 snapshot 可能已经发生，不能凭事件名字判断。
- [ ] 装备者伤害与队友伤害不能混淆 `ActorIndex`。
- [ ] 前台/后台、换人、重复装备、多人同名 buff 的覆盖规则。
- [ ] 首次触发、ICD 前一帧/当帧/后一帧、刷新与到期。
- [ ] 普攻/重击/E/Q/反应伤害/生成物是否正确过滤 `AttackTag`。
- [ ] 团队统计、元素/地区/阵营计数是在完整队伍初始化后读取。
- [ ] 动态属性和快照属性分别验证；modifier key 必须包含足够作用域，避免两把武器意外互相覆盖。

批量策略：每个 PR 以“一个共享 helper + 3~8 把同族武器”为上限；独特五星武器通常一把一个 PR。这样出错时可以单独回滚，也能让慢速审核并行进行。

## 6. 新增圣遗物

- [ ] 创建 `internal/artifacts/<key>/config.yml`，填 canonical `key` 和唯一 `set_id`。
- [ ] 编写 `Set`，实现 `SetIndex`、`GetCount`、`Init`、`NewSet`。
- [ ] 在 `init()` 中调用 `core.RegisterSetFunc(keys.<Set>, NewSet)`。
- [ ] 在 `pkg/core/keys/sets.go` 同步 canonical string 和 const；保持索引一致。
- [ ] 在 `pkg/shortcut/artifacts.go` 增加 canonical name/别名。
- [ ] 在 `pkg/simulation/imports.go` 增加 blank import。
- [ ] 运行 artifact pipeline，提交 UI JSON、资源映射和文档页；再用完整 pipeline 生成多语言名称数据。
- [ ] 手工补文档组件中的 Names，以及套装需要的 Params/Issues/AoE/Fields；pipeline 不会从 Go 被动实现中提取这些语义。
- [ ] 2 件套和 4 件套分别测试；1/3/5 件不应误触发。
- [ ] 需要队伍组成、Moonsign、武器或完整属性的逻辑放在 `Init()`。`NewSet` 在所有角色加入前执行，而装备 `Init()` 在全队建好后执行。
- [ ] 测试前后台、换人、持续时间、hitlag、叠层、同队多套、目标数量、反应类型过滤。
- [ ] 记录游戏不支持/模拟器不支持的描述分支，避免默认“全实现”。

## 7. Pipeline 和生成文件

### 7.1 环境

Pipeline 默认读取 `./pipeline/data`，也可以通过 `GENSHIN_DATA_REPO` 或 `--excels` 指向包含以下目录的数据根：

- `ExcelBinOutput/`
- `TextMap/`

角色至少依赖 Avatar、SkillDepot、Skill、Promote、ProudSkill、Fetter 数据；武器依赖 Weapon、WeaponCurve、WeaponPromote；圣遗物依赖 EquipAffix、ReliquarySet、Reliquary 数据。完整名称/文档/本地化还依赖各语言 TextMap。

### 7.2 命令

```powershell
# 建议显式固定数据路径
$env:GENSHIN_DATA_REPO = 'D:\data\genshin-data-pinned'

task pipeline-avatars
task pipeline-weapons
task pipeline-artifacts

# 合并前跑完整生成；只有四类全开时才生成完整多语言名称文件
task pipeline
```

- [ ] 永远不要手改标有 `Code generated` / `DO NOT EDIT` 的 Go 文件。
- [ ] 生成前后检查 `git diff`，确认没有因为数据源版本不同重写无关旧内容。
- [ ] PR 描述中写明外部数据 commit/hash 和生成命令。
- [ ] 把生成改动放入独立 commit，例如 `chore(pipeline): regenerate data for ...`。
- [ ] CI 增加“用固定数据快照运行 pipeline 后 `git diff --exit-code`”；数据不能放公开 CI 时，至少建立受控的 maintainer generation job。

当前自动化边界不一致：角色 key/import/ICD 会由 pipeline 生成，武器和圣遗物 key/shortcut/import 仍主要手工维护；角色和武器 config 中虽然声明了 `shortcuts`，generator 当前并未消费；文档组件数据也多为手工维护。`scripts/imports` 只生成 `out.txt` 且未接入 Taskfile。复活后的首批工程任务应统一这些注册数据的单一来源。

## 8. 合并前验证标准

```powershell
gofmt -w <本 PR 的手写 Go 文件>

# 与当前 CI 范围一致
go test ./internal/... ./pkg/gcs/... ./pkg/reactable/...

# 逐步扩成全仓；先记录并修复既有失败
go test ./...

task lint
go build ./cmd/gcsim
```

在权限受限的 Windows 环境中，可把缓存指向工作区，避免用户级缓存不可写：

```powershell
$env:GOCACHE = "$PWD\.cache\go-build"
$env:GOMODCACHE = "$PWD\.cache\go-mod"
```

- [ ] parser smoke test：canonical key 和每个 shortcut 都能解析。
- [ ] registry smoke test：新内容被 blank import，构造时不是 `unrecognized`。
- [ ] 固定 seed 的单目标与多目标样例。
- [ ] 按帧检查 sample log：snapshot、命中、反应、buff、伤害、附着、回调顺序。
- [ ] 对照独立计算器/实机/数据表，而不是只对照另一段 gcsim 代码。
- [ ] 改公式或反应时跑基准队伍差分，并解释每一个非零变化。
- [ ] 文档列出已知偏差和证据，不把“可运行”写成“精确完成”。

Definition of Done：代码、生成数据、注册、解析、文档、测试、证据、已知问题八项齐全；至少一名非作者做机制复核。

## 9. 当前仓库最值得先还的工程债

- [ ] 统一 Go 版本：`go.mod`、test action、lint workflow。
- [ ] 扩大 Go CI。当前注释明确写着全仓测试有既有问题，只运行了 `internal/...`、`pkg/gcs/...`、`pkg/reactable/...`。
- [ ] UI CI 真正运行 typecheck/lint/test/build；当前只 `yarn install --immutable`。
- [ ] 为角色、武器、圣遗物建立 table-driven 测试 helper；目前仅极少数角色专测，武器/圣遗物目录没有专测。
- [ ] 固定并记录 pipeline 数据版本，建立生成差分检查。
- [ ] 从 config 统一生成 key、shortcut、blank imports 和编辑器关键字，消除多处手工列表。
- [ ] 合并两个 `mode-gcsim.js` 副本；它们当前已经有轻微差异。
- [ ] 将 docs 的 Actions/Frames/Hitlag/AoE/Issues/Params 数据纳入有 schema 的校验。
- [ ] 清理或替换未接入构建的 `scripts/imports`。
- [ ] 给反应公式建立统一抽象。现有月反应倍率/特殊伤害分支分布在 `combat`、`reactable`、`enemy`，继续复制会增加漂移。
- [ ] 增加 `CODEOWNERS`、PR/issue 模板、`MAINTAINERS.md`、`GOVERNANCE.md`、`SECURITY.md`、发布 runbook。

## 10. PR 拆分和依赖顺序

推荐依赖图：

```text
PR-0 工具链/CI 基线
  |
  +--> PR-1 反应 A 规格与测试夹具 --> PR-2 反应 A 实现
  |
  +--> PR-3 反应 B 规格与测试夹具 --> PR-4 反应 B 实现
                                             |
                 +---------------------------+--------------------+
                 |                           |                    |
              角色 PR                  圣遗物 PR            武器批次 PR
```

规则：

- 一个 PR 只有一个清晰的回滚单位。
- 核心机制 PR 不夹带几十个内容包；内容 PR 不顺手重构全局事件系统。
- 共享 helper 先单独合并，调用它的内容 PR 才进入 ready 状态。
- 手写 commit 与生成 commit 分开；reviewer 首先看手写代码和测试。
- PR 描述必须列出依赖、验证证据、未实现项和对既有模拟结果的影响。
- 慢审核环境下保持较短分支，经常 rebase；不要让所有人堆在同一个长期 feature branch。

## 11. 如何和几个人一起复活项目

### 11.1 首选：在原项目内恢复维护能力

- [ ] 给现有 owner/maintainer 发一份**具体、低负担**的提案，不只说“我们想接手”：列出成员、可投入时间、负责模块、首月目标和风险控制。
- [ ] 请求分级权限：先 triage，再 review，再 merge；发布/基础设施权限最后移交。
- [ ] 提议 30 天试运行：你们负责分类 issue、复核 3~5 个小 PR、整理 release notes，owner 只做最终确认。
- [ ] 约定响应 SLA，例如普通 PR 7 天首次反馈、机制 PR 14 天；超时允许另一位 maintainer 接管。
- [ ] 明确谁有权合并 core/reaction、content、UI、infra；高风险模块要求两人批准。
- [ ] 请求 owner 写清域名、Cloudflare、GitHub Packages、数据库、bot、加密 key、发布签名和备份的所有权。不要通过聊天消息传长期密钥。
- [ ] 至少两人掌握 release，但生产 secrets 使用最小权限和环境审批。

### 11.2 如果原项目长期无法授权：透明维护社区 fork

MIT 许可证允许 fork、修改和发布，但要保留许可证和版权声明。

- [ ] fork 名称和 README 明确写“community-maintained fork / 非官方”，保留原项目 credit 和 upstream 链接。
- [ ] 保持 `upstream` remote，建立定期同步和冲突处理流程；能回馈原项目的修复继续提交 upstream PR。
- [ ] 尽量保持 GCSL、结果 schema 和 API 兼容；不兼容变更走版本化和迁移说明。
- [ ] 使用自己的域名、包名、容器 registry、Cloudflare 项目和 secrets，不能假设原项目部署资源可用。
- [ ] 发布包必须能从 tag 重现，附 commit、构建环境、checksum 和变更记录。
- [ ] 设定“何时正式启用 fork 发布”的客观门槛，例如连续 2 个版本周期无法合并关键兼容更新，而不是因一两个 PR 等待就立刻分裂社区。
- [ ] 定期公布与 upstream 的 ahead/behind、已回传 PR 和独有改动，避免静默形成不可维护的永久分叉。

### 11.3 团队角色，不要让所有人都只写代码

- **机制研究/实测**：帧数、附着、范围、快照、触发顺序、证据归档。
- **Core owner**：反应、事件、伤害、aura、target/gadget 等内核改动。
- **Content owner**：角色、武器、圣遗物实现和共享 helper。
- **Validation owner**：独立计算、固定 seed、实机对照、回归基准。
- **Release/infra owner**：CI、WASM、UI、文档、后端、部署和回滚。
- **Triage/reviewer**：拆 issue、控制 PR 大小、追依赖、维护 changelog。

每个条目至少“实现者 + 独立复核者”；反应和伤害公式需要两名非同一作者的 reviewer 更稳妥。

### 11.4 最小治理文件

- [ ] `MAINTAINERS.md`：模块 owner、权限、可联系状态。
- [ ] `GOVERNANCE.md`：决策、投票/共识、超时、冲突和移除失联维护者。
- [ ] `.github/CODEOWNERS`：core/reactable 至少两位 reviewer，内容包可分区 owner。
- [ ] PR template：证据、实现范围、生成命令、测试、已知偏差、breaking change。
- [ ] issue templates：mechanics research、bug、new content、regression。
- [ ] `SECURITY.md`：漏洞和密钥泄漏的私下报告渠道。
- [ ] release runbook：版本、tag、构建、部署、数据库迁移、回滚和事故沟通。
- [ ] roadmap：只承诺有 owner、有证据、有依赖顺序的内容。

## 12. 建议的前四周

### 第 1 周：盘点和获得信任

- [ ] 建群和每周 30 分钟同步；确定一个协调人，不设“终身项目领袖”。
- [ ] 完成内容矩阵、反应规格模板和贡献者能力分工。
- [ ] 联系原 owner，提交 30 天试运行提案。
- [ ] 对齐 Go/Node 版本，记录 CI 基线和 pipeline 数据版本。

### 第 2 周：让小 PR 快速流动

- [ ] 合并治理/PR 模板/测试 helper，不改游戏机制。
- [ ] 选择 2~3 把简单武器作为端到端演练，验证 config -> pipeline -> registry -> UI/docs -> test。
- [ ] 建立固定 seed regression configs 和生成 diff 检查。

### 第 3 周：反应内核

- [ ] 合并反应 A 的规格和测试，再合并实现。
- [ ] 反应 B 同样处理；若共享内核，先抽一个小而明确的公共层。
- [ ] 邀请原社区中熟悉元素论/帧数的人只做验证，不要求他们承担维护工作。

### 第 4 周：内容波次和首次发布演练

- [ ] 依赖反应的角色和圣遗物开始进入 PR 队列。
- [ ] 武器按同族批次推进，独特五星单独 PR。
- [ ] 从 tag 在干净环境构建 CLI/WASM/UI，演练回滚并发布候选版。
- [ ] 复盘吞吐量、review 等待时间、回归数，调整下一月承诺。

## 13. 不要做的事

- 不要把数据来源不明或未固定版本的生成文件大批提交。
- 不要仅凭技能文案猜测帧数、snapshot、ICD、范围和触发顺序。
- 不要用 DPS 看起来合理代替机制验证；多个错误可能互相抵消。
- 不要让同一个人研究、实现、批准并发布高风险反应改动。
- 不要复制 production secrets 到 fork，也不要把原域名/Discord 身份当作已经移交。
- 不要承诺一次补齐所有角色和几十把武器；先建立可重复的交付节奏。
- 不要因为 owner 审核慢就删除署名、制造对立或宣称接管官方项目；用透明治理和稳定发布赢得社区迁移。
