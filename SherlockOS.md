# SherlockOS — Technical Specification (v0.2)

## 1. 项目概述

SherlockOS 是一个基于"世界模型（World Model）+ 证据可靠性层级（Reliability Hierarchy）"的侦查辅助系统。它通过将不同来源的证据（Tier 0-3）映射到统一的空间与时间线上，利用 AI 发现时空矛盾（Spatial/Temporal Paradox），还原案件逻辑真相。

### 1.1 核心原则

1. **可靠性层级（Reliability Hierarchy）**：根据证据的客观性自动分配权重，从核心物理边界到主观证词。
2. **空间锚定（Spatial Anchoring）**：所有证据在 3D/2.5D 空间中拥有坐标或观察角度。
3. **时空辩证（Spatio-Temporal Deduction）**：利用硬性锚点（Hard Anchors）修正模糊证词（Soft Events）。

---

## 2. 证据可靠性层级 (Reliability Hierarchy)

| 级别 | 名称 | 定义 | 典型输入 | 逻辑权重 |
| :--- | :--- | :--- | :--- | :--- |
| **Tier 0** | **Environment** | 物理基础设施 / 不可穿透边界 | 平面图、静态扫描、3D 重建快照 | **100% / 物理底座** |
| **Tier 1** | **Ground Truth** | 原始视觉/录音记录 | CCTV、记录仪、关键帧视频 | **高 / 时空硬锚点** |
| **Tier 2** | **Electronic Logs** | 数字化触发记录 | 智能锁日志、Wi-Fi 连接、传感器 | **中 / 真实但存歧义** |
| **Tier 3** | **Testimonials** | 主观描述 | 目击者证词、嫌疑人辩解 | **低 / 主观概率** |

### 2.1 Tier 分类规则（按文件类型）

| Tier | 文件类型 | 后端 Job | 输出 |
|------|----------|----------|------|
| **Tier 0** | `.pdf`, `.e57`, `.dwg`, `.obj` | `reconstruction` | 3D Geometry / Proxy Geometry |
| **Tier 1** | `.mp4`, `.jpg`, `.png`, `.wav` | `reconstruction` + `scene_analysis` | Objects + Labels + Hard Anchors |
| **Tier 2** | `.json`, `.csv`, `.log` | Client-side parse | Timeline events |
| **Tier 3** | `.txt`, `.md`, `.docx` | `reasoning` | Motion paths + Claims |

---

## 3. 目标与非目标

### 3.1 目标（Goals）

* 在 hackathon 时间内做出可稳定 demo 的端到端闭环：
  * 上传证据 → Tier 自动分类 → 场景重建 → 时空推演 → 矛盾检测 → Paradox 高亮 → 导出报告
  * 输入证词 → 画像属性更新 → 画像生成 →（可选）与场景证据一致性打分联动
* UI 强调 "Reliability Hierarchy + Paradox Detection + Evidence/Reasoning/Simulation 三模式"
* 后续可推广：支持高并发读、实时推送、异步任务扩展、存储/CDN 分发

### 3.2 非目标（Non-goals）

* 不做"定罪工具"，只做"侦查辅助与假设生成"
* v0.2 不追求完美 3D 重建质量；允许 2.5D / Proxy Geometry 作为降级
* 不在 v0.2 做复杂多人协作编辑（但数据模型保留扩展能力）

---

## 4. 技术栈总览

### 4.1 Backend（Go）

* 语言：Go 1.22+
* HTTP 路由：chi（轻量）
* 实时推送：**Supabase Realtime**（优先，零代码订阅表变更）；自建 WebSocket 仅用于流式推理输出
* 数据访问：pgx（必选）；可选 sqlc（更工程化）
* 队列（hackathon 轻量版）：Redis Streams / Lists
  后续可升级：NATS / Kafka（推荐 NATS，轻快）
* 任务状态机：jobs 表 + 推送进度事件

### 4.2 Frontend

* Next.js (TypeScript)
* 3D：three.js（优先）
* 状态管理：Zustand
* 网络：REST + **Supabase Realtime**（订阅 commits/jobs 表变更）+ WS（仅推理流式输出）
* UI：Tailwind

### 4.3 Database & Storage：Supabase

* Postgres（主数据）
* Storage（上传原始扫描图 / 生成图 / mesh / pointcloud 等）
* Realtime（订阅 commits / jobs 变化，用于 timeline/进度实时刷新）
* Auth（可选：后续推广做用户体系与 RLS）

### 4.4 模型与能力集成（Google Ecosystem Focus）

#### 4.4.1 核心 AI 引擎：Gemini 3 Pro (Action Era Engine)

| 能力 | 说明 |
|------|------|
| **1M Context Window** | 摄入全案卷宗、长时间视频与海量传感器日志，无需常规 RAG |
| **Thought Signatures & Thinking Levels** | 记录推理深度，实现自主路径规划与自我纠错 |
| **Spatio-Temporal Video Understanding** | 识别 CCTV 画面中的因果关系（如：玻璃破碎与嫌疑人动作的物理联系） |

**用途**：
- 轨迹推理与 Paradox 检测
- Perspective Validation（视线遮挡验证）
- Hard Anchor Mapping（模糊证词时间校准）
- 证词结构化属性提取

#### 4.4.2 场景重建 / Proxy Geometry

| 能力 | 说明 |
|------|------|
| **Scene Reconstruction Engine** | 从多视角还原 Tier 0 环境 |
| **Proxy Geometry** | 将物体转化为简单几何体（Box/Cylinder），供 Gemini 运行空间逻辑 |
| **2.5D Overlay** | 在原始侦查照片上叠加深度值，支持虚拟测距 |

降级策略：
- 完整重建不可用时，退回 Proxy Geometry + 粗定位
- 全部失败时，返回 Mock SceneGraph + 高 uncertainty

#### 4.4.3 Image Gen：Nano Banana Pro

| 场景 | 模型 | 说明 |
|------|------|------|
| 画像迭代（快速） | Nano Banana | 低延迟，适合多轮编辑 |
| 高保真输出（报告） | Nano Banana Pro | 支持 4K，**Localized Paint-to-Edit** 精准修正 |

**用途**：
- 嫌疑人画像生成/精准局部编辑
- Evidence board 渲染图、对比图、报告插图

**API 文档**：https://ai.google.dev/gemini-api/docs/image-generation

#### 4.4.4 未来演进：D4RT

系统架构完全兼容 Google **D4RT (Dynamic 4D Reconstruction & Tracking)**。未来通过简单的 Worker 接入，即可实现更极致的 4D 动态追踪与视频-空间无缝映射。

---

## 5. 侦查辅助核心逻辑

### 5.1 Chronological Analysis（时空映射）

系统利用 **Hard Anchor Mapping** 机制：

1. **硬锚点**: CCTV 显示嫌疑人 22:05 进入。
2. **模糊证词**: 目击者称"开门后大约几分钟听到响动"。
3. **演绎**: 系统将"响动"映射为 22:08±60s，并由 Gemini 检查该时间内其他传感器的并发记录。

### 5.2 Discrepancy Detection（矛盾发现）

* **Sightline Paradox（视线矛盾）**：AI 计算证明目击者位置与宣称看到的物体间存在墙壁（Tier 0）。
* **Temporal Paradox（时间悖论）**：嫌疑人称在此处，但 Tier 2 日志显示其设备在彼处。

---

## 6. 工作流：The Sherlock Pipeline

```
Ingestion → Model → Deduction → Simulation
```

### 6.1 Ingestion（证据摄入与空间转换）

* 摄入多模态证据并根据 Tier (0-3) 自动分配初始权重。
* **Spatial Extraction**: 利用 D4RT (Tier 1) 或 Gemini (Tier 3) 同步提取初步的空间轨迹；Gemini 从证词中即时推算目击者的**运动路径 (Motion Path)**。

### 6.2 Model（世界建模）

* 通过重建引擎构建静态物理环境与 **Proxy Geometry（代理几何体）**。
* 建立空间基准坐标系，确定"硬性不可穿越"边界。

### 6.3 Deduction（逻辑推演与矛盾探测）

* **Perspective Validation**: 在 3D 空间中模拟目击者的视点（POV），利用 Proxy Geometry 验证宣称的观察是否受物理遮挡影响。
* **Discrepancy Identification**: 对比 Tier 3 (Testimonials) 与 Tier 0/1/2，高亮显示不合逻辑的"时空矛盾"节点（如：视线被墙阻挡，或时间线冲突）。

### 6.4 Simulation（真实还原仿真）

* 将所有经过验证的"硬事实"与校准后的"软证据"结合，进行 4D 动态全景仿真。
* 生成"最优拟合"的案件经过复现。

---

## 7. 核心产品功能规格（Feature Spec）

### 7.1 Timeline（版本控制式事件流）

**定义：** 每一次用户输入（上传扫描/输入证词/手动编辑/运行推理）都会生成一个 commit，形成可回放的 timeline。

**必须支持：**

* commit 列表（时间、类型、摘要、作者）
* commit diff（哪些对象新增/更新/删除、哪些证据置信度变化）
* 选中 commit → 场景与面板回滚到对应版本（只读回放）

### 7.2 SceneGraph（世界状态，Single Source of Truth）

SceneGraph 是整个系统的核心数据结构，存储在 `scene_snapshots.scenegraph` (JSONB)。

#### 7.2.1 JSON Schema 定义

```typescript
interface SceneGraph {
  version: string;                    // Schema 版本，如 "1.0.0"
  bounds: BoundingBox;                // 场景整体包围盒
  objects: SceneObject[];             // 可交互实体列表
  evidence: EvidenceCard[];           // 证据卡列表
  constraints: Constraint[];          // 约束条件
  uncertainty_regions?: UncertaintyRegion[];  // 不确定区域（可选）
  paradoxes?: Paradox[];              // 检测到的时空矛盾（可选）
}

interface SceneObject {
  id: string;                         // UUID
  type: ObjectType;                   // 枚举：见下方
  label: string;                      // 显示名称
  pose: Pose;                         // 位置与朝向
  bbox: BoundingBox;                  // 包围盒（Proxy Geometry）
  mesh_ref?: string;                  // Storage key（可选）
  state: ObjectState;                 // 枚举：visible | occluded | suspicious | removed
  evidence_ids: string[];             // 关联的证据 ID
  confidence: number;                 // 0-1，重建置信度
  tier: 0 | 1 | 2 | 3;               // 证据可靠性层级
  source_commit_ids: string[];        // 贡献该对象的 commit IDs
  metadata?: Record<string, unknown>; // 扩展字段
}

type ObjectType =
  | "furniture" | "door" | "window" | "wall"
  | "evidence_item" | "weapon" | "footprint" | "bloodstain"
  | "vehicle" | "person_marker" | "other";

interface Pose {
  position: [number, number, number]; // [x, y, z] 米
  rotation: [number, number, number, number]; // 四元数 [w, x, y, z]
  scale?: [number, number, number];   // 默认 [1, 1, 1]
}

interface BoundingBox {
  min: [number, number, number];
  max: [number, number, number];
}

interface EvidenceCard {
  id: string;
  object_ids: string[];               // 关联对象
  title: string;
  description: string;
  confidence: number;                 // 0-1
  tier: 0 | 1 | 2 | 3;               // 来源 Tier
  sources: EvidenceSource[];          // 来源列表
  conflicts?: EvidenceSource[];       // 冲突来源
  created_at: string;                 // ISO 8601
}

interface EvidenceSource {
  type: "upload" | "witness" | "inference" | "electronic_log";
  commit_id: string;
  description?: string;
  credibility?: number;               // 0-1，仅 witness 类型
  tier: 0 | 1 | 2 | 3;               // 来源 Tier
}

interface Constraint {
  id: string;
  type: ConstraintType;
  description: string;
  params: Record<string, unknown>;    // 类型相关参数
  confidence: number;
}

type ConstraintType =
  | "door_direction"      // params: { object_id, direction: "inward" | "outward" }
  | "passable_area"       // params: { polygon: [x,y][] }
  | "height_range"        // params: { min_cm, max_cm }
  | "time_window"         // params: { start_iso, end_iso }
  | "impassable_boundary" // params: { object_ids: string[] } — Tier 0 hard boundary
  | "custom";

interface Paradox {
  id: string;
  type: "sightline" | "temporal";
  severity: "low" | "medium" | "high";
  description: string;
  evidence_refs: ParadoxEvidenceRef[];
  spatial_location?: [number, number, number]; // 矛盾发生的空间位置
  time_range?: { start: string; end: string }; // 矛盾涉及的时间范围
}

interface ParadoxEvidenceRef {
  evidence_id: string;
  tier: 0 | 1 | 2 | 3;
  role: "anchor" | "contradicted";    // anchor = 高 Tier 锚点，contradicted = 被矛盾的低 Tier
}

interface UncertaintyRegion {
  id: string;
  bbox: BoundingBox;
  level: "low" | "medium" | "high";
  reason: string;
}
```

#### 7.2.2 示例数据

```json
{
  "version": "1.0.0",
  "bounds": { "min": [0, 0, 0], "max": [10, 3, 8] },
  "objects": [
    {
      "id": "obj_001",
      "type": "door",
      "label": "主入口门",
      "pose": { "position": [5, 0, 0], "rotation": [1, 0, 0, 0] },
      "bbox": { "min": [4.5, 0, -0.1], "max": [5.5, 2.2, 0.1] },
      "state": "visible",
      "evidence_ids": ["ev_003"],
      "confidence": 0.92,
      "tier": 0,
      "source_commit_ids": ["commit_abc123"]
    }
  ],
  "evidence": [
    {
      "id": "ev_003",
      "object_ids": ["obj_001"],
      "title": "门锁损坏痕迹",
      "description": "门锁有明显撬痕，金属变形",
      "confidence": 0.85,
      "tier": 1,
      "sources": [{ "type": "upload", "commit_id": "commit_abc123", "tier": 1 }]
    }
  ],
  "constraints": [
    {
      "id": "con_001",
      "type": "door_direction",
      "description": "主入口门向内开",
      "params": { "object_id": "obj_001", "direction": "inward" },
      "confidence": 0.95
    },
    {
      "id": "con_002",
      "type": "impassable_boundary",
      "description": "北墙为承重墙，不可穿越",
      "params": { "object_ids": ["obj_wall_north"] },
      "confidence": 1.0
    }
  ],
  "paradoxes": [
    {
      "id": "paradox_001",
      "type": "sightline",
      "severity": "high",
      "description": "目击者 A 声称看到嫌疑人从窗户进入，但 Tier 0 重建显示其站立位置与窗户间有承重墙遮挡",
      "evidence_refs": [
        { "evidence_id": "ev_wall_north", "tier": 0, "role": "anchor" },
        { "evidence_id": "ev_witness_a", "tier": 3, "role": "contradicted" }
      ],
      "spatial_location": [5, 1.5, 3]
    }
  ]
}
```

### 7.3 视图模式（View Modes）

系统通过三个核心模式引导侦查员完成案件分析：

| 模式 | 定义 | 核心功能 |
| :--- | :--- | :--- |
| **Evidence Mode** | 静态映射与标记 | 展示 **Evidence Archive**。查看重建的物理现场（Tier 0）与同步的物证资产。 |
| **Reasoning Mode** | 动态推演与对比 | **多轨时间轴 (Multi-track Timeline)**。时间轴按 **Distinct Scene Volumes (独立 3D 场景)** 与 **Stakeholders (人/目击者)** 分行。点击时间段 Block 即可在 3D 视图中触发跨场景的 **Motion Path Ghost（轨迹残影）**。 |
| **Simulation Mode** | 4D 全景还原 | **真相回放**。拖动时间轴进行 4D 全景仿真，自动检测并高亮 **Paradox Alerts**。 |

### 7.4 假设分支（Hypothesis Branches）

假设分支允许在不同约束条件下进行对比分析：

| 功能 | 说明 |
|------|------|
| 创建分支 | 从任意 commit 创建 Branch A/B |
| 修改约束 | 分支可覆盖约束（如"门向内开 vs 向外开"） |
| 对比分析 | 分支间对比轨迹结果与置信度评分 |
| 合并采纳 | 将分支结论合并为主线（v0.3 扩展） |

**数据模型**：分支通过 `branches` 表记录，commit 通过 `branch_id` 关联。

### 7.5 Suspect Profiling（嫌疑人画像侧写）

采取"两层结构"避免"纯 prompt 生成像编故事"：

1. **属性层（Structured Attributes）**

* 年龄段、身高区间、体型、肤色区间、发型、眼镜/胡子、显著特征等
* 每个属性带：
  * probability（置信度）
  * supporting_sources（证词/证据来源 + Tier）
  * conflict_sources（冲突来源）

2. **图像层（Rendered Portrait）**

* 由 Nano Banana Pro 基于属性层生成/编辑图像
* **Localized Paint-to-Edit**：精准修正特定区域（如只改发型）
* 支持并排多版本（例如"有胡子 vs 无胡子"）

### 7.6 Export Report（导出案件报告）

一键导出：

* 场景截图（俯视/关键区域）
* 证据列表（按 Tier 分组，卡片摘要）
* 检测到的 Paradox 汇总（含证据引用与严重度）
* 轨迹候选与解释（含证据引用）
* 嫌疑人画像与属性层摘要
* 不确定区域与"下一步建议"

输出形式：

* HTML（hackathon 最快）
* 或 PDF（后续再做）

---

## 8. 数据模型（Supabase Postgres）

> 说明：下面给出推荐表结构（字段可按实现裁剪）。核心是 **append-only commits + current snapshot**。

### 8.1 表结构

> **约定**：所有表使用 `gen_random_uuid()` 生成主键，时间戳使用 `timestamptz` 并默认 `now()`。

#### 枚举类型定义（Postgres ENUM）

```sql
CREATE TYPE commit_type AS ENUM (
  'upload_scan',
  'witness_statement',
  'manual_edit',
  'reconstruction_update',
  'profile_update',
  'reasoning_result',
  'export_report'
);

CREATE TYPE job_type AS ENUM (
  'reconstruction',
  'imagegen',
  'reasoning',
  'profile',
  'export'
);

CREATE TYPE job_status AS ENUM (
  'queued',
  'running',
  'done',
  'failed',
  'canceled'
);

CREATE TYPE asset_kind AS ENUM (
  'scan_image',
  'generated_image',
  'mesh',
  'pointcloud',
  'portrait',
  'report'
);

CREATE TYPE evidence_tier AS ENUM (
  'tier_0',
  'tier_1',
  'tier_2',
  'tier_3'
);
```

#### cases

```sql
CREATE TABLE cases (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title           text NOT NULL,
  description     text,
  created_by      uuid,  -- 关联 auth.users（后续启用）
  created_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT cases_title_length CHECK (char_length(title) <= 200)
);

CREATE INDEX idx_cases_created_at ON cases(created_at DESC);
```

#### commits（Timeline，append-only）

```sql
CREATE TABLE commits (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id           uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  parent_commit_id  uuid REFERENCES commits(id),
  branch_id         uuid REFERENCES branches(id),
  type              commit_type NOT NULL,
  summary           text NOT NULL,
  payload           jsonb NOT NULL DEFAULT '{}',  -- diff/patch、引用资产、证据变化
  created_by        uuid,
  created_at        timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT commits_summary_length CHECK (char_length(summary) <= 500)
);

CREATE INDEX idx_commits_case_id ON commits(case_id, created_at DESC);
CREATE INDEX idx_commits_branch_id ON commits(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX idx_commits_payload_job_id ON commits((payload->>'job_id')) WHERE payload->>'job_id' IS NOT NULL;
```

#### branches

```sql
CREATE TABLE branches (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id         uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  name            text NOT NULL,
  base_commit_id  uuid NOT NULL REFERENCES commits(id),
  created_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT branches_name_length CHECK (char_length(name) <= 100),
  CONSTRAINT branches_unique_name UNIQUE (case_id, name)
);

CREATE INDEX idx_branches_case_id ON branches(case_id);
```

#### scene_snapshots（Current State 快照）

```sql
CREATE TABLE scene_snapshots (
  case_id     uuid PRIMARY KEY REFERENCES cases(id) ON DELETE CASCADE,
  commit_id   uuid NOT NULL REFERENCES commits(id),
  scenegraph  jsonb NOT NULL,  -- 结构见 7.2 节 SceneGraph Schema
  updated_at  timestamptz NOT NULL DEFAULT now()
);
```

#### suspect_profiles

```sql
CREATE TABLE suspect_profiles (
  case_id             uuid PRIMARY KEY REFERENCES cases(id) ON DELETE CASCADE,
  commit_id           uuid NOT NULL REFERENCES commits(id),
  attributes          jsonb NOT NULL DEFAULT '{}',  -- 结构化属性层
  portrait_asset_key  text,  -- Storage key
  updated_at          timestamptz NOT NULL DEFAULT now()
);
```

#### jobs

```sql
CREATE TABLE jobs (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id          uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  type             job_type NOT NULL,
  status           job_status NOT NULL DEFAULT 'queued',
  progress         integer NOT NULL DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
  input            jsonb NOT NULL,
  output           jsonb,
  error            text,
  idempotency_key  text,  -- 客户端幂等键
  retry_count      integer NOT NULL DEFAULT 0,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT jobs_idempotency_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_jobs_case_id ON jobs(case_id, created_at DESC);
CREATE INDEX idx_jobs_status ON jobs(status) WHERE status IN ('queued', 'running');
CREATE INDEX idx_jobs_heartbeat ON jobs(updated_at) WHERE status = 'running';
```

#### assets

```sql
CREATE TABLE assets (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id      uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  kind         asset_kind NOT NULL,
  tier         evidence_tier,  -- 证据可靠性层级
  storage_key  text NOT NULL,
  metadata     jsonb DEFAULT '{}',
  created_at   timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT assets_storage_key_unique UNIQUE (storage_key)
);

CREATE INDEX idx_assets_case_id ON assets(case_id);
CREATE INDEX idx_assets_kind ON assets(case_id, kind);
CREATE INDEX idx_assets_tier ON assets(case_id, tier);
```

### 8.2 RLS（后续推广必需）

* cases：仅 owner / team 可读写
* commits：同 case 权限继承
* storage：按 case 前缀隔离

hackathon 可先关闭 RLS，保留结构以便迁移。

---

## 9. API 规格（Go Backend）

### 9.1 REST API（v1）

#### 9.1.1 通用响应格式

```typescript
// 成功响应
interface ApiResponse<T> {
  success: true;
  data: T;
  meta?: { cursor?: string; total?: number };
}

// 错误响应
interface ApiError {
  success: false;
  error: {
    code: ErrorCode;
    message: string;
    details?: Record<string, unknown>;
  };
}

type ErrorCode =
  | "INVALID_REQUEST"      // 400 - 请求格式错误
  | "UNAUTHORIZED"         // 401 - 未认证
  | "FORBIDDEN"            // 403 - 无权限
  | "NOT_FOUND"            // 404 - 资源不存在
  | "CONFLICT"             // 409 - 资源冲突（如重复创建）
  | "RATE_LIMITED"         // 429 - 请求过频
  | "JOB_FAILED"           // 500 - 任务执行失败
  | "MODEL_UNAVAILABLE"    // 503 - 模型服务不可用
  | "INTERNAL_ERROR";      // 500 - 内部错误
```

#### 9.1.2 接口清单与示例

---

**POST /v1/cases** - 创建案件

```bash
curl -X POST /v1/cases \
  -H "Content-Type: application/json" \
  -d '{"title": "案件 A-2025-001", "description": "某住宅入室盗窃"}'
```

响应 `201 Created`:
```json
{
  "success": true,
  "data": {
    "id": "case_abc123",
    "title": "案件 A-2025-001",
    "description": "某住宅入室盗窃",
    "created_at": "2025-09-15T10:30:00Z"
  }
}
```

---

**GET /v1/cases/{caseId}/snapshot** - 获取当前 SceneGraph

响应 `200 OK`:
```json
{
  "success": true,
  "data": {
    "case_id": "case_abc123",
    "commit_id": "commit_xyz789",
    "scenegraph": { /* SceneGraph JSON，见 7.2 节 */ },
    "updated_at": "2025-09-15T11:00:00Z"
  }
}
```

---

**POST /v1/cases/{caseId}/upload-intent** - 获取上传预签名 URL（含自动 Tier 分类）

> **选型决策**：使用 **Go 后端签名**（非 Supabase SDK 直传），便于统一权限控制与审计。

```bash
curl -X POST /v1/cases/case_abc123/upload-intent \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"filename": "scan_001.jpg", "content_type": "image/jpeg", "size_bytes": 2048000},
      {"filename": "floorplan.pdf", "content_type": "application/pdf", "size_bytes": 1500000}
    ]
  }'
```

响应 `200 OK`:
```json
{
  "success": true,
  "data": {
    "upload_batch_id": "batch_123",
    "intents": [
      {
        "filename": "scan_001.jpg",
        "tier": 1,
        "storage_key": "cases/case_abc123/scans/batch_123/scan_001.jpg",
        "presigned_url": "https://xxx.supabase.co/storage/v1/object/sign/...",
        "expires_at": "2025-09-15T11:30:00Z"
      },
      {
        "filename": "floorplan.pdf",
        "tier": 0,
        "storage_key": "cases/case_abc123/scans/batch_123/floorplan.pdf",
        "presigned_url": "https://xxx.supabase.co/storage/v1/object/sign/...",
        "expires_at": "2025-09-15T11:30:00Z"
      }
    ]
  }
}
```

---

**POST /v1/cases/{caseId}/jobs** - 创建任务

```bash
curl -X POST /v1/cases/case_abc123/jobs \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem_xyz789" \
  -d '{
    "type": "reconstruction",
    "input": {
      "scan_asset_keys": [
        "cases/case_abc123/scans/batch_123/scan_001.jpg",
        "cases/case_abc123/scans/batch_123/floorplan.pdf"
      ],
      "camera_poses": null
    }
  }'
```

响应 `202 Accepted`:
```json
{
  "success": true,
  "data": {
    "job_id": "job_def456",
    "type": "reconstruction",
    "status": "queued",
    "progress": 0,
    "created_at": "2025-09-15T11:05:00Z"
  }
}
```

错误示例 `409 Conflict`（重复提交）:
```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "Job with idempotency key already exists",
    "details": { "existing_job_id": "job_def456" }
  }
}
```

---

**GET /v1/jobs/{jobId}** - 查询任务状态

响应 `200 OK`:
```json
{
  "success": true,
  "data": {
    "job_id": "job_def456",
    "type": "reconstruction",
    "status": "running",
    "progress": 45,
    "output": null,
    "error": null,
    "created_at": "2025-09-15T11:05:00Z",
    "updated_at": "2025-09-15T11:07:30Z"
  }
}
```

---

**POST /v1/cases/{caseId}/witness-statements** - 提交证词（Tier 3）

```bash
curl -X POST /v1/cases/case_abc123/witness-statements \
  -H "Content-Type: application/json" \
  -d '{
    "statements": [
      {
        "source_name": "目击者 A",
        "content": "嫌疑人约 175cm，短发，戴眼镜",
        "credibility": 0.8
      }
    ]
  }'
```

响应 `201 Created`:
```json
{
  "success": true,
  "data": {
    "commit_id": "commit_wit001",
    "type": "witness_statement",
    "tier": 3,
    "profile_job_id": "job_profile_001"
  }
}
```

---

* `GET /v1/cases/{caseId}` - 案件详情
* `GET /v1/cases/{caseId}/timeline?cursor=xxx&limit=50` - commit 列表（分页）
* `POST /v1/cases/{caseId}/branches` - 创建假设分支
* `POST /v1/cases/{caseId}/reasoning` - 触发推理（内部创建 reasoning job）
* `POST /v1/cases/{caseId}/export` - 触发导出（内部创建 export job）

### 9.2 实时推送方案

> **选型决策**：采用 **Supabase Realtime 为主 + 自建 WS 为辅** 的混合方案。

#### 9.2.1 Supabase Realtime（主要方案）

用于订阅表变更，零后端代码：

```typescript
// 前端订阅 commits 表
const channel = supabase
  .channel('case-timeline')
  .on('postgres_changes', {
    event: 'INSERT',
    schema: 'public',
    table: 'commits',
    filter: `case_id=eq.${caseId}`
  }, (payload) => {
    // 新 commit 到达，刷新 timeline
    handleNewCommit(payload.new);
  })
  .on('postgres_changes', {
    event: 'UPDATE',
    schema: 'public',
    table: 'jobs',
    filter: `case_id=eq.${caseId}`
  }, (payload) => {
    // job 进度更新
    handleJobUpdate(payload.new);
  })
  .subscribe();
```

#### 9.2.2 自建 WebSocket（仅推理流式输出）

当 Reasoning Worker 需要流式返回思考过程时使用：

频道：`/v1/ws/reasoning?jobId=...`

事件类型：

| 事件 | Payload | 说明 |
|------|---------|------|
| `thinking_chunk` | `{text: string}` | 推理思考过程片段 |
| `paradox_detected` | `{paradox: Paradox}` | 实时检测到的 Paradox |
| `trajectory_partial` | `{index: number, segment: TrajectorySegment}` | 轨迹片段 |
| `complete` | `{commit_id: string}` | 推理完成 |
| `error` | `{code: string, message: string}` | 错误 |

#### 9.2.3 方案对比

| 场景 | 使用方案 | 原因 |
|------|----------|------|
| Timeline 更新 | Supabase Realtime | 零代码，直接订阅 commits 表 |
| Job 进度 | Supabase Realtime | 零代码，直接订阅 jobs 表 |
| 推理流式输出 | 自建 WS | 需要流式传输模型输出 + Paradox 实时推送 |
| 画像生成进度 | Supabase Realtime | 轮询 jobs 表即可 |

---

## 10. 任务编排与 Worker 规格

### 10.1 通用 Job 生命周期

#### 状态机

```
queued → running → done
           ↓
        failed ← (可重试)
           ↓
       canceled (用户取消)
```

progress：0–100（worker 每 5 秒或每步回写）

#### 幂等策略（防止重复写入）

1. **请求级幂等**：客户端在 `POST /jobs` 时携带 `Idempotency-Key` Header
   - 后端在 `jobs` 表增加 `idempotency_key` 字段（UNIQUE 约束）
   - 相同 key 的请求返回已存在的 job，不重复创建

2. **Worker 级幂等**：每个 job 产出的 commit 携带 `job_id` 引用
   - 回写前检查：`SELECT 1 FROM commits WHERE payload->>'job_id' = $1`
   - 若已存在，跳过写入，标记 job 为 done

3. **模型调用幂等**：
   - 对于确定性输入（如同一批图片），缓存模型输出 hash
   - 重试时优先检查缓存

#### 重试策略

```go
type RetryConfig struct {
    MaxAttempts     int           // 默认 3
    InitialInterval time.Duration // 默认 2s
    MaxInterval     time.Duration // 默认 30s
    Multiplier      float64       // 默认 2.0
}
```

| 错误类型 | 处理方式 |
|----------|----------|
| 模型 API 5xx / 超时 | 指数退避重试，最多 3 次 |
| 模型 API 4xx（非 429） | 立即失败，不重试 |
| 429 Rate Limit | 按 Retry-After 等待后重试 |
| Worker crash | Job 保持 `running`，由 heartbeat 检测后重新分配 |
| 输入数据无效 | 立即失败，记录详细 error |

#### Heartbeat 与僵尸任务检测

- Worker 每 30 秒更新 `jobs.updated_at`
- 调度器检测 `status = running AND updated_at < NOW() - INTERVAL '2 minutes'` 的任务
- 僵尸任务重置为 `queued`（重试次数 +1）或标记 `failed`（超过 MaxAttempts）

### 10.2 Reconstruction Worker（Scene Reconstruction Engine）

#### 输入 Schema

```typescript
interface ReconstructionInput {
  case_id: string;
  scan_asset_keys: string[];           // 多视角图片 Storage keys
  tiers: Record<string, 0 | 1>;       // 每个 asset 的 Tier（0=平面图, 1=照片）
  camera_poses?: CameraPose[];         // 可选：相机位姿
  depth_maps?: string[];               // 可选：深度图 keys
  existing_scenegraph?: SceneGraph;    // 增量更新时传入
}

interface CameraPose {
  asset_key: string;
  intrinsics: { fx: number; fy: number; cx: number; cy: number };
  extrinsics: { rotation: number[]; translation: number[] };
}
```

#### 输出 Schema

```typescript
interface ReconstructionOutput {
  objects: SceneObjectProposal[];
  proxy_geometry: ProxyGeometrySet;     // Proxy Geometry 集合
  mesh_asset_key?: string;             // 整体 mesh（glb/obj）
  pointcloud_asset_key?: string;       // 点云
  impassable_boundaries: string[];     // Tier 0 不可穿越边界 object IDs
  uncertainty_regions: UncertaintyRegion[];
  processing_stats: {
    input_images: number;
    detected_objects: number;
    processing_time_ms: number;
  };
}

interface SceneObjectProposal {
  id: string;
  action: "create" | "update" | "remove";
  object: Partial<SceneObject>;
  confidence: number;
  source_images: string[];             // 贡献该对象的图片 keys
}

interface ProxyGeometrySet {
  boxes: ProxyBox[];
  cylinders: ProxyCylinder[];
}

interface ProxyBox {
  object_id: string;
  center: [number, number, number];
  dimensions: [number, number, number]; // width, height, depth
  rotation: [number, number, number, number]; // quaternion
}

interface ProxyCylinder {
  object_id: string;
  center: [number, number, number];
  radius: number;
  height: number;
}
```

#### 降级策略

| 情况 | 降级方案 |
|------|----------|
| 完整重建不可用 | 退回 Proxy Geometry + 粗定位 |
| 无相机位姿 | 使用自动位姿估计 |
| GPU 内存不足 | 分批处理图片（每批 4 张） |
| 全部失败 | 返回 Mock SceneGraph + 高 uncertainty |

### 10.3 Profile Worker（证词 → 属性层）

输入：

* witness statements（Tier 3，带来源/可信度）
* existing attributes

输出（profile_update commit + suspect_profiles 更新）：

* attributes（每项概率、支持/冲突来源、Tier 标注）
* 触发 imagegen job（Nano Banana Pro 画像生成）

### 10.4 ImageGen Worker（Nano Banana Pro）

#### 输入 Schema

```typescript
interface ImageGenInput {
  case_id: string;
  gen_type: "portrait" | "evidence_board" | "comparison" | "report_figure";

  // portrait 类型
  portrait_attributes?: SuspectAttributes;
  reference_image_key?: string;        // 上一版画像
  edit_region?: EditRegion;            // Localized Paint-to-Edit 区域

  // evidence_board / comparison 类型
  object_ids?: string[];
  layout?: "grid" | "timeline" | "comparison";

  // 通用
  resolution: "1k" | "2k" | "4k";
  style_prompt?: string;
}

interface EditRegion {
  region_type: "face" | "hair" | "glasses" | "facial_hair" | "clothing" | "custom";
  mask_points?: [number, number][];    // 自定义区域多边形
  edit_prompt: string;                 // 该区域的编辑指令
}

interface SuspectAttributes {
  age_range: { min: number; max: number; confidence: number };
  height_range_cm: { min: number; max: number; confidence: number };
  build: { value: "slim" | "average" | "heavy"; confidence: number };
  skin_tone: { value: string; confidence: number };
  hair: { style: string; color: string; confidence: number };
  facial_hair?: { type: string; confidence: number };
  glasses?: { type: string; confidence: number };
  distinctive_features: Array<{ description: string; confidence: number }>;
}
```

#### 输出 Schema

```typescript
interface ImageGenOutput {
  asset_key: string;                   // 生成图片 Storage key
  thumbnail_key: string;               // 缩略图 key
  resolution: { width: number; height: number };
  model_used: "nano-banana" | "nano-banana-pro";
  generation_time_ms: number;
  cost_usd: number;
}
```

### 10.5 Reasoning Worker（Gemini 3 Pro — Deduction Stage）

#### 输入 Schema

```typescript
interface ReasoningInput {
  case_id: string;
  scenegraph: SceneGraph;              // 含 Proxy Geometry 和 Tier 标注
  evidence_by_tier: {                  // 按 Tier 分组的证据
    tier_0: EvidenceCard[];
    tier_1: EvidenceCard[];
    tier_2: EvidenceCard[];
    tier_3: EvidenceCard[];
  };
  branch_id?: string;                  // 分支推理
  constraints_override?: Constraint[]; // 覆盖/新增约束
  thinking_budget?: number;            // Thought Signatures depth
  max_trajectories?: number;           // Top-K，默认 3
}
```

#### 输出 Schema

```typescript
interface ReasoningOutput {
  trajectories: Trajectory[];
  paradoxes: Paradox[];                // 检测到的时空矛盾
  hard_anchor_mappings: HardAnchorMapping[]; // 模糊事件→硬锚点校准
  uncertainty_areas: UncertaintyRegion[];
  next_step_suggestions: Suggestion[];
  thinking_summary?: string;           // 可选：思考过程摘要
  model_stats: {
    thinking_tokens: number;
    output_tokens: number;
    latency_ms: number;
  };
}

interface Trajectory {
  id: string;
  rank: number;                        // 1 = 最可能
  overall_confidence: number;          // 0-1
  segments: TrajectorySegment[];
}

interface TrajectorySegment {
  id: string;
  from_position: [number, number, number];
  to_position: [number, number, number];
  waypoints?: [number, number, number][];  // 可视化用
  time_estimate?: { start: string; end: string };
  evidence_refs: EvidenceRef[];
  confidence: number;
  explanation: string;                 // 短句解释
}

interface EvidenceRef {
  evidence_id: string;
  object_id?: string;
  tier: 0 | 1 | 2 | 3;
  relevance: "supports" | "contradicts" | "neutral";
  weight: number;                      // 对该段置信度的贡献
}

interface HardAnchorMapping {
  soft_event_id: string;               // Tier 3 模糊事件
  hard_anchor_id: string;              // Tier 0/1 硬锚点
  calibrated_time: string;             // 校准后的时间 ISO 8601
  uncertainty_seconds: number;         // ±误差范围
  explanation: string;
}

interface Suggestion {
  type: "collect_evidence" | "verify_constraint" | "interview" | "analyze";
  description: string;
  priority: "high" | "medium" | "low";
  related_object_ids?: string[];
}
```

#### Prompt 模板

```python
REASONING_PROMPT = """
你是一个专业的案件分析助手。基于以下分层证据数据，执行时空推演并检测矛盾。

## 证据可靠性层级
- Tier 0 (Environment): 物理不可穿越边界，100% 可信
- Tier 1 (Ground Truth): 视觉/录音记录，高可信度硬锚点
- Tier 2 (Electronic Logs): 数字化记录，真实但可能存在歧义
- Tier 3 (Testimonials): 主观描述，需要验证

## 场景数据（含 Proxy Geometry）
{scenegraph_json}

## 分层证据
{evidence_by_tier_json}

## 约束条件
{constraints_json}

## 任务
1. **Hard Anchor Mapping**: 将 Tier 3 模糊事件锚定到 Tier 0/1 硬锚点
2. **Perspective Validation**: 验证 Tier 3 目击者视线是否被 Tier 0 结构遮挡
3. **Discrepancy Detection**: 识别 Tier 3 与 Tier 0/1/2 的时空矛盾
4. 推断 Top-{max_trajectories} 条可能的移动轨迹
5. 为每段轨迹提供证据引用（含 Tier 标注）和置信度
6. 标注不确定区域
7. 给出下一步侦查建议

## 输出格式
严格按照以下 JSON Schema 输出：
{output_schema}
"""
```

---

## 11. UI/UX 规格（Next.js）

### 11.1 主界面布局

```
┌─────────────────────────────────────────────────────────────────────┐
│ Header: Logo | Case Title | [Evidence] [Reasoning] [Simulation]    │
│                                              | Jobs | Export       │
├──────────┬────────────────────────────────────────────┬─────────────┤
│ Sidebar  │  3D Scene Viewer                           │ Right Panel │
│ ──────── │  ┌───────────────────────────────────────┐ │ ─────────── │
│ Evidence │  │  Proxy Geometry objects               │ │ Context:    │
│ Archive  │  │  Motion Path Ghosts                   │ │ - Suspect   │
│          │  │  Paradox Alert markers (red)          │ │ - Evidence  │
│ [Tier 0] │  │  POV camera (simulation mode)         │ │ - Reasoning │
│ [Tier 1] │  └───────────────────────────────────────┘ │             │
│ [Tier 2] │  [Evidence] [Reasoning] [Simulation]       │ Paradoxes   │
│ [Tier 3] │                                            │ Confidence  │
│          ├────────────────────────────────────────────┴─────────────┤
│ ──────── │  Multi-track Timeline                                    │
│ Drop     │  ├─ Scene Volumes: [Room A] [Room B] [Exterior]         │
│ Zone     │  ├─ Stakeholders:  [Suspect] [Witness A] [Witness B]    │
│          │  └─ Paradox Alerts: ⚠️ ──── ⚠️ ────────── ⚠️            │
└──────────┴──────────────────────────────────────────────────────────┘
```

### 11.2 关键交互

* Hover 对象：高亮 + 提示"由哪些 commits/哪个 Tier 贡献"
* 点击时间段 Block：触发跨场景 Motion Path Ghost
* Paradox Alert 点击：高亮矛盾双方证据 + 展示 Tier 对比
* 分支切换：A/B 假设并排对比（评分 + 差异点）
* 一键导出：生成报告并弹出下载链接

---

## 12. 安全与合规（最低要求）

* 明确免责声明：输出为"侦查辅助假设"，显示不确定性
* 资产访问：使用短期签名 URL，避免公开 bucket
* 后续推广：开启 RLS + Auth + 审计日志

---

## 13. 可观测性与运维

* Go 服务：
  * structured logs（JSON）
  * request tracing（OpenTelemetry，可选）
* Worker：
  * 每个 job 的 step 日志 + failure reason
* Dashboard（hackathon 可简化）：
  * jobs 列表、失败重试、吞吐统计

---

## 14. 部署建议（hackathon → 产品化）

### hackathon（最快）

* Frontend：Vercel 或任意容器
* Go API：单容器（Render/Fly/Cloud Run/自建）
* Workers：1–2 个容器（按任务类型分）
* Redis：托管或容器
* Supabase：托管项目 + Storage + Realtime

### 推广阶段（扛流量）

* Go API 横向扩展（无状态）
* CDN 缓存静态资源与图片
* Worker autoscaling（按队列长度扩）
* 队列升级到 NATS（或 Kafka）
* 引入成本控制：结果缓存 + 增量更新优先

---

## 15. Demo 展示脚本（3 分钟建议）

1. 设定案件 + 打开空白 Case（0:00–0:20）
2. Upload 平面图（Tier 0）→ Proxy Geometry 出现（0:20–0:50）
3. Upload CCTV 关键帧（Tier 1）→ Hard Anchor 标记 + 新对象（0:50–1:20）
4. 输入 2 条证词（Tier 3，含矛盾）→ 属性层变化 + 画像并排（1:20–1:50）
5. 触发 Deduction → **Sightline Paradox 高亮** + Temporal Paradox 检测（1:50–2:30）
6. Simulation Mode → 4D 回放 + Paradox Alert 自动暂停（2:30–2:50）
7. 导出报告 + "下一步建议"（2:50–3:00）

---

## 16. 技术演进 (Future-proofing)

系统架构完全兼容 Google **D4RT (Dynamic 4D Reconstruction & Tracking)**。未来通过简单的 Worker 接入，即可实现更极致的 4D 动态追踪与视频-空间无缝映射，深度融入 Google 侦查生态。

---

## 17. 关键实现优先级（MVP Checklist）

**P0（必须有）**

* Case + Timeline（commits）
* 上传证据（自动 Tier 分类）+ reconstruction job（Proxy Geometry）
* SceneGraph 快照 + 3D/2.5D 可视化（Proxy Geometry 渲染）
* Reasoning job + Paradox Detection（Sightline/Temporal Paradox）
* 证词输入（Tier 3）+ 属性层更新 + Nano Banana Pro 画像生成
* Export HTML（或至少生成一个 shareable summary 页面）

**P1（强加分）**

* Simulation Mode（4D 回放 + Paradox Alert 自动暂停）
* Branching hypotheses A/B
* Uncertainty heatmap（对象级也行）
* Hard Anchor Mapping 可视化（模糊事件→校准时间）
* 缓存与增量更新

---

## 18. 假设与约束（明确写在 README 里）

* 扫描输入形式：多张图片（优先）/视频帧（可选）
* 若没有相机位姿/深度：重建质量下降，但可降级为"Proxy Geometry + 粗定位"
* Nano Banana Pro 的出图结果不作为 truth，只作为展示与报告资产
* 推理依赖结构化 SceneGraph + Tier 分层证据，缺失信息将产生 uncertainty
* 所有 Paradox 为"辅助假设"，不构成法律判断

---

## 附录 A：外部依赖与 API 参考

| 服务 | 文档链接 | 用途 |
|------|----------|------|
| Gemini API | https://ai.google.dev/gemini-api/docs | 核心推理引擎 |
| Nano Banana Pro | https://ai.google.dev/gemini-api/docs/image-generation | 画像生成 + Paint-to-Edit |
| Gemini Thinking | https://ai.google.dev/gemini-api/docs/thinking | Thought Signatures |
| D4RT (Future) | TBD | 4D 动态追踪 |
| Supabase | https://supabase.com/docs | 数据库 + 存储 + 实时订阅 |

---

## 附录 B：Frontend Implementation Plan

> **Updated: 2026-02**

### Implementation Priorities

#### P0 - MVP Core (Evidence Mode)

| Feature | Description | Components |
|---------|-------------|------------|
| **File Drop + Auto-Classification** | Drag files → detect type → assign Tier (0-3) | `DropZone`, `FileClassifier` |
| **Upload + Job Trigger** | Tier 0-1 → reconstruction, Tier 3 → witness API | `useUpload` hook, API integration |
| **Job Progress UI** | Real-time status via Supabase | `JobProgress` component |
| **Real SceneGraph → 3D** | Fetch snapshot → render Proxy Geometry | `SceneViewer` update |
| **Timeline Commits** | Version history with diff view | `CommitTimeline` component |

#### P1 - Reasoning Mode

| Feature | Description | Components |
|---------|-------------|------------|
| **Multi-track Timeline** | Tracks by Scene Volume + Stakeholder | `MultiTrackTimeline` |
| **Paradox Alert Markers** | Sightline / Temporal paradox display | `ParadoxMarker` |
| **Motion Path Ghosts** | Cross-scene trajectory overlays | `MotionPathGhost` |
| **Discrepancy Highlights** | Tier 3 vs Tier 0-1-2 contradictions | `DiscrepancyOverlay` |

#### P2 - Simulation Mode

| Feature | Description | Components |
|---------|-------------|------------|
| **4D Replay** | Timeline scrubber with animated playback | `SimulationPlayer` |
| **POV Simulation** | Witness perspective camera | `POVCamera` |
| **Perspective Validation** | Occlusion detection from POV | `OcclusionChecker` |
| **Paradox Auto-Pause** | Pause playback at paradox timestamps | `ParadoxPause` |

### Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                          EVIDENCE MODE                               │
├─────────────────────────────────────────────────────────────────────┤
│  Drop File → Classify Tier → Upload → Trigger Job → Update Scene    │
│                                                                      │
│  Tier 0:   reconstruction → Proxy Geometry → Impassable boundaries  │
│  Tier 1:   reconstruction → Objects + Hard Anchors                  │
│  Tier 2:   parse JSON/CSV → Timeline events                        │
│  Tier 3:   witness API → Motion path proposals                     │
├─────────────────────────────────────────────────────────────────────┤
│                         REASONING MODE                               │
├─────────────────────────────────────────────────────────────────────┤
│  Compare Tier 3 claims ↔ Tier 0-1-2 facts                           │
│  Detect: 视线遮挡 Sightline Paradox (line-of-sight blocked)         │
│  Detect: 时间线冲突 Temporal Paradox (timeline conflicts)            │
│  Output: Paradox nodes highlighted in scene + timeline              │
├─────────────────────────────────────────────────────────────────────┤
│                         SIMULATION MODE                              │
├─────────────────────────────────────────────────────────────────────┤
│  4D Replay: Drag timeline → animate verified trajectories            │
│  POV Simulation: Switch to witness camera → validate sightline       │
│  Auto-pause at Paradox Alerts                                        │
└─────────────────────────────────────────────────────────────────────┘
```

---

*文档版本：v0.2 | 最后更新：2026-02-09*
