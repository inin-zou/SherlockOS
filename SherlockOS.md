# SherlockOS — Technical Specification (v0.1)

## 1. 项目概述

SherlockOS 是一个基于“世界模型（World Model）+ 时间线（Timeline）+ 可解释推理（Explainable Reasoning）”的侦查辅助 Web App，用于：

1. **Crime Scene 世界状态迭代**
   用户上传扫描空间的多视角图片（或视频帧/扫描输出），系统逐步重建并优化案发现场的结构化世界状态（SceneGraph / Scene State），并展示可交互场景视图（2.5D/3D）。

2. **嫌疑人画像侧写（Suspect Profiling）**
   用户输入多条目击者证词（可带来源与可信度），系统迭代更新嫌疑人“结构化属性层”，再生成/编辑嫌疑人画像图。

3. **轨迹推理（Movement Trajectory Reasoning）**
   当场景/证据达到一定完整度后，调用 Gemini 2.5 Flash（带 Thinking 模式）对嫌疑人可能的运动轨迹进行多假设推理，并提供证据引用、置信度与可解释链路。

> 核心原则：**Scene State（结构化世界状态）是 Single Source of Truth**。所有推理与展示都从 Scene State 派生，而不是直接从图片“脑补”。

---

## 2. 目标与非目标

### 2.1 目标（Goals）

* 在 hackathon 时间内做出可稳定 demo 的端到端闭环：

  * 上传扫描图 → 场景状态更新 → 证据卡片 → 轨迹推理 → 解释模式 → 导出报告
  * 输入证词 → 画像属性更新 → 画像生成 →（可选）与场景证据一致性打分联动
* UI 强调 “Timeline + Evidence/Reasoning 分离 + Explain Mode”
* 后续可推广：支持高并发读、实时推送、异步任务扩展、存储/CDN 分发

### 2.2 非目标（Non-goals）

* 不做“定罪工具”，只做“侦查辅助与假设生成”
* v0.1 不追求完美 3D 重建质量；允许 2.5D / proxy geometry 作为降级
* 不在 v0.1 做复杂多人协作编辑（但数据模型保留扩展能力）

---

## 3. 命名与模块边界

### 3.1 模块边界

* **Go Backend（Control Plane）**：状态、timeline、实时推送、鉴权、任务编排、API
* **AI Workers（Data/Model Plane）**：异步调用模型（重建/生成/推理），产出结构化结果回写
* **Supabase**：Postgres（数据）+ Storage（资产）+ Realtime（订阅）+ Auth（可选）
* **Frontend**：Next.js + three.js（或 Babylon.js）实现 Timeline + 场景交互 + Explain Mode

---

## 4. 技术栈总览

### 4.1 Backend（最终选用 Go）

* 语言：Go 1.22+
* HTTP 路由：chi / gin（二选一，建议 chi，轻量）
* 实时推送：**Supabase Realtime**（优先，零代码订阅表变更）；自建 WebSocket 仅用于流式推理输出
* 数据访问：pgx（必选）；可选 sqlc（更工程化）
* 队列（hackathon 轻量版）：Redis Streams / Lists
  后续可升级：NATS / Kafka（推荐 NATS，轻快）
* 任务状态机：jobs 表 + 推送进度事件

### 4.2 Frontend

* Next.js (TypeScript)
* 3D：three.js（优先）或 Babylon.js
* 状态管理：Zustand
* 网络：REST + **Supabase Realtime**（订阅 commits/jobs 表变更）+ WS（仅推理流式输出）
* UI：Tailwind（或任意组件库）

### 4.3 Database & Storage：Supabase

* Postgres（主数据）
* Storage（上传原始扫描图 / 生成图 / mesh / pointcloud 等）
* Realtime（订阅 commits / jobs 变化，用于 timeline/进度实时刷新）
* Auth（可选：后续推广做用户体系与 RLS）

### 4.4 模型与能力集成

> **可用性说明**：以下列出生产优先方案 + 降级备选，确保 hackathon 可跑通。

#### 4.4.1 场景重建 / 世界一致性（Tencent Hunyuan 系列）

| 能力 | 生产优先 | 降级备选 | 接入方式 |
|------|----------|----------|----------|
| 场景几何重建 | **HunyuanWorld-Mirror** | HunyuanWorld-Voyager (RGBD) | 自部署 / Replicate |
| 世界初始化 | **HunyuanWorld-1.0** | Mock SceneGraph | 自部署（4090 可跑 lite 版） |
| 物品 3D 资产 | **Hunyuan3D-2.1** | Hunyuan3D-2mini (5GB VRAM) | Replicate API / Tencent Cloud (20次/天免费) |
| 图像理解 | HunyuanImage-3.0 | Gemini 2.5 Flash Vision | 自部署 / AI/ML API |

**API 参考**：
- Hunyuan3D-2: https://replicate.com/tencent/hunyuan3d-2/api
- HunyuanWorld-Mirror: https://github.com/Tencent-Hunyuan/HunyuanWorld-Mirror
- HunyuanImage-3.0: https://github.com/Tencent-Hunyuan/HunyuanImage-3.0

#### 4.4.2 Image Gen（Nano Banana 系列，Google）

| 场景 | 模型 | Model ID | 说明 |
|------|------|----------|------|
| 快速生成（画像迭代） | **Nano Banana** | `gemini-2.5-flash-image` | 低延迟，适合多轮编辑 |
| 高保真输出（报告/打印） | **Nano Banana Pro** | `gemini-3-pro-image-preview` | 支持 4K，文字渲染清晰 |

**用途**：
- 嫌疑人画像生成/编辑
- Evidence board 渲染图、对比图、报告插图

**定价参考**（Google AI）：
- Nano Banana: ~$0.04/image (1K)
- Nano Banana Pro: ~$0.134/image (2K), ~$0.24/image (4K)

**API 文档**：https://ai.google.dev/gemini-api/docs/image-generation

#### 4.4.3 轨迹推理（Gemini 2.5，强调可解释）

| 模型 | Model ID | 特点 |
|------|----------|------|
| **Gemini 2.5 Flash** (推荐) | `gemini-2.5-flash` | 快速推理，支持 thinking budget |
| Gemini 2.5 Pro | `gemini-2.5-pro` | 更强推理，成本较高 |

**Thinking 配置**：
```python
from google import genai
client = genai.Client(api_key="GEMINI_API_KEY")
response = client.models.generate_content(
    model="gemini-2.5-flash",
    contents=reasoning_prompt,
    config=genai.types.GenerateContentConfig(
        thinking_config=genai.types.ThinkingConfig(
            thinking_budget=8192  # 0-24576, -1 for dynamic
        )
    )
)
```

**输入**：SceneGraph + Evidence + Constraints
**输出**：轨迹候选（Top-K）、每段证据引用、置信度、不确定性区域、下一步建议

**API 文档**：https://ai.google.dev/gemini-api/docs/thinking

---

## 5. 核心产品功能规格（Feature Spec）

### 5.1 Timeline（版本控制式事件流）

**定义：** 每一次用户输入（上传扫描/输入证词/手动编辑/运行推理）都会生成一个 commit，形成可回放的 timeline。

**必须支持：**

* commit 列表（时间、类型、摘要、作者）
* commit diff（哪些对象新增/更新/删除、哪些证据置信度变化）
* 选中 commit → 场景与面板回滚到对应版本（只读回放）

### 5.2 SceneGraph（世界状态，Single Source of Truth）

SceneGraph 是整个系统的核心数据结构，存储在 `scene_snapshots.scenegraph` (JSONB)。

#### 5.2.1 JSON Schema 定义

```typescript
interface SceneGraph {
  version: string;                    // Schema 版本，如 "1.0.0"
  bounds: BoundingBox;                // 场景整体包围盒
  objects: SceneObject[];             // 可交互实体列表
  evidence: EvidenceCard[];           // 证据卡列表
  constraints: Constraint[];          // 约束条件
  uncertainty_regions?: UncertaintyRegion[];  // 不确定区域（可选）
}

interface SceneObject {
  id: string;                         // UUID
  type: ObjectType;                   // 枚举：见下方
  label: string;                      // 显示名称
  pose: Pose;                         // 位置与朝向
  bbox: BoundingBox;                  // 包围盒
  mesh_ref?: string;                  // Storage key（可选）
  state: ObjectState;                 // 枚举：visible | occluded | suspicious | removed
  evidence_ids: string[];             // 关联的证据 ID
  confidence: number;                 // 0-1，重建置信度
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
  sources: EvidenceSource[];          // 来源列表
  conflicts?: EvidenceSource[];       // 冲突来源
  created_at: string;                 // ISO 8601
}

interface EvidenceSource {
  type: "upload" | "witness" | "inference";
  commit_id: string;
  description?: string;
  credibility?: number;               // 0-1，仅 witness 类型
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
  | "custom";

interface UncertaintyRegion {
  id: string;
  bbox: BoundingBox;
  level: "low" | "medium" | "high";
  reason: string;
}
```

#### 5.2.2 示例数据

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
      "sources": [{ "type": "upload", "commit_id": "commit_abc123" }]
    }
  ],
  "constraints": [
    {
      "id": "con_001",
      "type": "door_direction",
      "description": "主入口门向内开",
      "params": { "object_id": "obj_001", "direction": "inward" },
      "confidence": 0.95
    }
  ]
}
```

### 5.3 视图模式（View Modes）

> **术语统一**：使用中文「证据模式」「推理模式」「解释模式」，或英文 Evidence/Reasoning/Explain Mode。

| 模式 | 英文 | 核心功能 |
|------|------|----------|
| **证据模式** | Evidence Mode | 展示现场对象、证据标注、冲突提示、不确定性热力图 |
| **推理模式** | Reasoning Mode | 展示轨迹、假设分支、推理解释链路、评分对比 |
| **解释模式** | Explain Mode | 可交互的推理过程溯源 |

### 5.4 解释模式（Explain Mode）

* 轨迹分段（Segment）可点击
* 点击某段 → 高亮相关对象/证据卡 + 自动定位 timeline 中贡献该证据的 commit
* 每段显示：

  * 证据引用列表（Evidence IDs）
  * 置信度（0–1）
  * 推断规则摘要（短句）

### 5.5 假设分支（Hypothesis Branches）

> **术语统一**：使用「假设分支」而非 Branching Hypotheses。

假设分支允许在不同约束条件下进行对比分析：

| 功能 | 说明 |
|------|------|
| 创建分支 | 从任意 commit 创建 Branch A/B |
| 修改约束 | 分支可覆盖约束（如"门向内开 vs 向外开"） |
| 对比分析 | 分支间对比轨迹结果与置信度评分 |
| 合并采纳 | 将分支结论合并为主线（v0.2 扩展） |

**数据模型**：分支通过 `branches` 表记录，commit 通过 `branch_id` 关联。

### 5.6 Suspect Profiling（嫌疑人画像侧写）

采取“两层结构”避免“纯 prompt 生成像编故事”：

1. **属性层（Structured Attributes）**

* 年龄段、身高区间、体型、肤色区间、发型、眼镜/胡子、显著特征等
* 每个属性带：

  * probability（置信度）
  * supporting_sources（证词/证据来源）
  * conflict_sources（冲突来源）

2. **图像层（Rendered Portrait）**

* 由 Nano Banana 基于属性层生成/编辑图像
* 支持并排多版本（例如“有胡子 vs 无胡子”）

### 5.7 Suspect–Scene Fit Score（联动加分项，推荐做）

* 从属性层提取可与场景一致性相关的特征（如身高、步幅、惯用手、是否跛行）
* 与场景中可测线索（鞋印间距、门把手高度、可达区域）比对
* 输出一个 Fit Score（0–100）并解释原因

### 5.8 Export Report（导出案件报告）

一键导出：

* 场景截图（俯视/关键区域）
* 证据列表（卡片摘要）
* 轨迹候选与解释（含证据引用）
* 嫌疑人画像与属性层摘要
* 不确定区域与“下一步建议”

输出形式：

* HTML（hackathon 最快）
* 或 PDF（后续再做）

---

## 6. 系统架构

### 6.1 高层架构图（文字版）

1. Frontend 上传扫描图片 → 获取预签名 URL → 直传 Supabase Storage
2. Frontend 调用 Go API 创建 `job(reconstruction)`
3. Go 将 job 写入 DB + 投递队列
4. Reconstruction Worker 拉取 job：

   * 调用 HunyuanWorld-Mirror / HunyuanWorld-1.0
   * 输出对象提案、位姿更新、（可选）mesh/pointcloud 链接
   * 回写 SceneGraph（生成 commit）
5. Frontend 订阅 commits/jobs（Realtime 或 WS）→ timeline 自动更新
6. 用户输入证词 → 生成 commit → 触发 profile job
7. Profile Worker 结构化属性更新 → 触发 Nano Banana 画像生成 → 回写画像 asset
8. 推理阶段：Reasoning Worker 调用 Gemini 2.5 Flash 输出轨迹 + 解释 → 回写 commit
9. Export：Go 生成 report artifact（HTML/PDF），存 Storage

---

## 7. 数据模型（Supabase Postgres）

> 说明：下面给出推荐表结构（字段可按实现裁剪）。核心是 **append-only commits + current snapshot**。

### 7.1 表结构

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
  scenegraph  jsonb NOT NULL,  -- 结构见 5.2 节 SceneGraph Schema
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
  storage_key  text NOT NULL,
  metadata     jsonb DEFAULT '{}',
  created_at   timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT assets_storage_key_unique UNIQUE (storage_key)
);

CREATE INDEX idx_assets_case_id ON assets(case_id);
CREATE INDEX idx_assets_kind ON assets(case_id, kind);
```

### 7.2 RLS（后续推广必需）

* cases：仅 owner / team 可读写
* commits：同 case 权限继承
* storage：按 case 前缀隔离

hackathon 可先关闭 RLS，保留结构以便迁移。

---

## 8. API 规格（Go Backend）

### 8.1 REST API（v1）

#### 8.1.1 通用响应格式

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

#### 8.1.2 接口清单与示例

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
    "scenegraph": { /* SceneGraph JSON，见 5.2 节 */ },
    "updated_at": "2025-09-15T11:00:00Z"
  }
}
```

---

**POST /v1/cases/{caseId}/upload-intent** - 获取上传预签名 URL

> **选型决策**：使用 **Go 后端签名**（非 Supabase SDK 直传），便于统一权限控制与审计。

```bash
curl -X POST /v1/cases/case_abc123/upload-intent \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"filename": "scan_001.jpg", "content_type": "image/jpeg", "size_bytes": 2048000},
      {"filename": "scan_002.jpg", "content_type": "image/jpeg", "size_bytes": 1850000}
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
        "storage_key": "cases/case_abc123/scans/batch_123/scan_001.jpg",
        "presigned_url": "https://xxx.supabase.co/storage/v1/object/sign/...",
        "expires_at": "2025-09-15T11:30:00Z"
      },
      {
        "filename": "scan_002.jpg",
        "storage_key": "cases/case_abc123/scans/batch_123/scan_002.jpg",
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
        "cases/case_abc123/scans/batch_123/scan_002.jpg"
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

**POST /v1/cases/{caseId}/witness-statements** - 提交证词

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

### 8.2 实时推送方案

> **选型决策**：采用 **Supabase Realtime 为主 + 自建 WS 为辅** 的混合方案。

#### 8.2.1 Supabase Realtime（主要方案）

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

#### 8.2.2 自建 WebSocket（仅推理流式输出）

当 Reasoning Worker 需要流式返回思考过程时使用：

频道：`/v1/ws/reasoning?jobId=...`

事件类型：

| 事件 | Payload | 说明 |
|------|---------|------|
| `thinking_chunk` | `{text: string}` | 推理思考过程片段 |
| `trajectory_partial` | `{index: number, segment: TrajectorySegment}` | 轨迹片段 |
| `complete` | `{commit_id: string}` | 推理完成 |
| `error` | `{code: string, message: string}` | 错误 |

#### 8.2.3 方案对比

| 场景 | 使用方案 | 原因 |
|------|----------|------|
| Timeline 更新 | Supabase Realtime | 零代码，直接订阅 commits 表 |
| Job 进度 | Supabase Realtime | 零代码，直接订阅 jobs 表 |
| 推理流式输出 | 自建 WS | 需要流式传输模型输出 |
| 画像生成进度 | Supabase Realtime | 轮询 jobs 表即可 |

---

## 9. 任务编排与 Worker 规格

### 9.1 通用 Job 生命周期

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

### 9.2 Reconstruction Worker（HunyuanWorld-Mirror）

#### 输入 Schema

```typescript
interface ReconstructionInput {
  case_id: string;
  scan_asset_keys: string[];           // 多视角图片 Storage keys
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
  mesh_asset_key?: string;             // 整体 mesh（glb/obj）
  pointcloud_asset_key?: string;       // 点云
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
```

#### 降级策略

| 情况 | 降级方案 |
|------|----------|
| HunyuanWorld-Mirror 不可用 | 切换 HunyuanWorld-Voyager |
| 无相机位姿 | 使用 Mirror 的自动位姿估计 |
| GPU 内存不足 | 分批处理图片（每批 4 张） |
| 全部失败 | 返回 Mock SceneGraph + 高 uncertainty |

### 9.3 Profile Worker（证词 → 属性层）

输入：

* witness statements（带来源/可信度）
* existing attributes

输出（profile_update commit + suspect_profiles 更新）：

* attributes（每项概率、支持/冲突来源）
* 触发 imagegen job（Nano Banana 画像生成）

### 9.4 ImageGen Worker（Nano Banana / Nano Banana Pro）

#### 输入 Schema

```typescript
interface ImageGenInput {
  case_id: string;
  gen_type: "portrait" | "evidence_board" | "comparison" | "report_figure";

  // portrait 类型
  portrait_attributes?: SuspectAttributes;
  reference_image_key?: string;        // 上一版画像

  // evidence_board / comparison 类型
  object_ids?: string[];
  layout?: "grid" | "timeline" | "comparison";

  // 通用
  resolution: "1k" | "2k" | "4k";      // 1k 用 Nano Banana，2k/4k 用 Pro
  style_prompt?: string;
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

#### 模型选择策略

| 场景 | 模型 | 原因 |
|------|------|------|
| 画像迭代（快速预览） | Nano Banana | 低延迟，便宜 |
| 最终画像（报告用） | Nano Banana Pro | 高清，文字清晰 |
| Evidence board | Nano Banana Pro | 需要渲染多个元素 |
| 对比图 | Nano Banana | 快速对比即可 |

### 9.5 Reasoning Worker（Gemini 2.5 Flash）

#### 输入 Schema

```typescript
interface ReasoningInput {
  case_id: string;
  scenegraph: SceneGraph;
  branch_id?: string;                  // 分支推理
  constraints_override?: Constraint[]; // 覆盖/新增约束
  thinking_budget?: number;            // 0-24576，默认 8192
  max_trajectories?: number;           // Top-K，默认 3
}
```

#### 输出 Schema

```typescript
interface ReasoningOutput {
  trajectories: Trajectory[];
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
  relevance: "supports" | "contradicts" | "neutral";
  weight: number;                      // 对该段置信度的贡献
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
你是一个专业的案件分析助手。基于以下场景数据，推断嫌疑人可能的移动轨迹。

## 场景数据
{scenegraph_json}

## 约束条件
{constraints_json}

## 任务
1. 分析场景中的证据关系
2. 推断 Top-{max_trajectories} 条可能的移动轨迹
3. 为每段轨迹提供证据引用和置信度
4. 标注不确定区域
5. 给出下一步侦查建议

## 输出格式
严格按照以下 JSON Schema 输出：
{output_schema}
"""
```

---

## 10. UI/UX 规格（Next.js）

### 10.1 主界面布局（推荐）

* 左：Timeline（commits 列表 + diff）
* 中：Scene View（2.5D/3D，点击对象显示 Evidence Card）
* 右：Reasoning Panel / Profile Panel（模式切换）
* 顶：Mode Toggle（Evidence / Reasoning / Explain）

### 10.2 关键交互

* Hover 对象：高亮 + 提示“由哪些 commits/哪张图贡献”
* 点击轨迹段：高亮路径段 + 关联证据卡闪烁 + timeline 自动定位
* 分支切换：A/B 假设并排对比（评分 + 差异点）
* 一键导出：生成报告并弹出下载链接

---

## 11. 安全与合规（最低要求）

* 明确免责声明：输出为“侦查辅助假设”，显示不确定性
* 资产访问：使用短期签名 URL，避免公开 bucket
* 后续推广：开启 RLS + Auth + 审计日志

---

## 12. 可观测性与运维

* Go 服务：

  * structured logs（JSON）
  * request tracing（OpenTelemetry，可选）
* Worker：

  * 每个 job 的 step 日志 + failure reason
* Dashboard（hackathon 可简化）：

  * jobs 列表、失败重试、吞吐统计

---

## 13. 部署建议（hackathon → 产品化）

### hackathon（最快）

* Frontend：Vercel（Vercel）或任意容器
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

## 14. Demo 展示脚本（3 分钟建议）

1. 设定案件 + 打开空白 Case（0:00–0:20）
2. Upload Scan #1 → 粗场景出现（0:20–1:00）
3. Upload Scan #2 → 新证据/对象出现 + 冲突提示（1:00–1:30）
4. Reasoning → 轨迹 Top2 + Explain Mode 点段看证据链（1:30–2:20）
5. 输入 2 条证词（含矛盾）→ 属性层变化 + 画像并排（2:20–2:50）
6. 导出报告 + “下一步建议”（2:50–3:00）

---

## 15. 关键实现优先级（MVP Checklist）

**P0（必须有）**

* Case + Timeline（commits）
* 上传扫描图（Storage）+ reconstruction job（哪怕返回 mock objects）
* SceneGraph 快照 + 3D/2.5D 可视化
* Reasoning job + Explain Mode（轨迹段→证据卡）
* 证词输入 + 属性层更新 + Nano Banana 画像生成
* Export HTML（或至少生成一个 shareable summary 页面）

**P1（强加分）**

* Branching hypotheses A/B
* Uncertainty heatmap（对象级也行）
* Suspect–Scene Fit Score
* 缓存与增量更新

---

## 15.1 Frontend Implementation Plan (Phase 5)

> **Updated: 2026-02**

### Evidence Tier System

| Tier | Category | File Types | Backend Job | Output |
|------|----------|------------|-------------|--------|
| **Tier 0** | Environment | `.pdf`, `.e57`, `.dwg`, `.obj` | `reconstruction` | 3D Geometry/Mapping |
| **Tier 1** | Ground Truth | `.mp4`, `.jpg`, `.png`, `.wav` | `reconstruction` + `scene_analysis` | Objects + Labels |
| **Tier 2** | Electronic Logs | `.json`, `.csv`, `.log` | Client-side parse | Timeline events |
| **Tier 3** | Testimonials | `.txt`, `.md`, `.docx` | `reasoning` | Motion paths + Claims |

### Implementation Priorities

#### P0 - MVP Core (Evidence Mode)

| Feature | Description | Components |
|---------|-------------|------------|
| **File Drop + Auto-Classification** | Drag files → detect type → assign tier | `DropZone`, `FileClassifier` |
| **Upload + Job Trigger** | Tier 0-1 → reconstruction, Tier 3 → witness API | `useUpload` hook, API integration |
| **Job Progress UI** | Real-time status via Supabase | `JobProgress` component |
| **Real SceneGraph → 3D** | Fetch snapshot → render Three.js objects | `SceneViewer` update |
| **Timeline Commits** | Version history with diff view | `CommitTimeline` component |

#### P1 - Enhanced Features

| Feature | Description | Components |
|---------|-------------|------------|
| **Suspect Profile Panel** | Attributes + Portrait display | `SuspectPanel` |
| **Discrepancy Highlights** | Tier 3 vs Tier 0-1-2 contradictions | `DiscrepancyOverlay` |
| **Evidence Detail Modal** | Click item → full details | `EvidenceModal` |
| **Witness Statement Input** | Form to add testimonials | `WitnessForm` |

#### P2 - Advanced Features (Simulation & Reasoning)

| Feature | Description | Components |
|---------|-------------|------------|
| **Text → Motion Path** | Description → trajectory inference | `MotionPathGenerator` |
| **Video → Motion** | CCTV analysis → movement extraction | `VideoAnalyzer` |
| **POV Simulation** | Witness perspective camera | `POVCamera` |
| **Perspective Validation** | Occlusion detection from POV | `OcclusionChecker` |

### Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                          EVIDENCE MODE                               │
├─────────────────────────────────────────────────────────────────────┤
│  Drop File → Classify Tier → Upload → Trigger Job → Update Scene    │
│                                                                      │
│  Tier 0-1: reconstruction → SceneGraph objects → 3D render          │
│  Tier 2:   parse JSON/CSV → Timeline events → Track visualization   │
│  Tier 3:   witness API → Motion path proposals → Trajectory render  │
├─────────────────────────────────────────────────────────────────────┤
│                         SIMULATION MODE                              │
├─────────────────────────────────────────────────────────────────────┤
│  Text Input → Reasoning Job → Trajectory → Animate in 3D            │
│  Video Upload → Scene Analysis → Extract Motion → Path render       │
├─────────────────────────────────────────────────────────────────────┤
│                          REASONING MODE                              │
├─────────────────────────────────────────────────────────────────────┤
│  Compare Tier 3 claims ↔ Tier 0-1-2 facts                           │
│  Detect: 视线遮挡 (line-of-sight blocked)                            │
│  Detect: 时间线冲突 (timeline conflicts)                              │
│  Output: Highlighted discrepancy nodes in scene + timeline          │
└─────────────────────────────────────────────────────────────────────┘
```

### UI Layout (Updated)

```
┌─────────────────────────────────────────────────────────────────────┐
│ Header: Logo | Search | Case Tabs | [Jobs ●] | Export               │
├──────────┬────────────────────────────────────────────┬─────────────┤
│ Sidebar  │  3D Scene Viewer                           │ Right Panel │
│ ──────── │  ┌───────────────────────────────────────┐ │ ─────────── │
│ Evidence │  │  Real SceneGraph objects              │ │ Context:    │
│ Archive  │  │  Real trajectories from reasoning     │ │ - Suspect   │
│          │  │  Discrepancy highlights (red)         │ │ - Evidence  │
│ [Tier 0] │  │  POV camera (reasoning mode)          │ │ - Reasoning │
│ [Tier 1] │  └───────────────────────────────────────┘ │             │
│ [Tier 2] │  [Evidence] [Simulation] [Reasoning]       │ Attributes  │
│ [Tier 3] │                                            │ Confidence  │
│          ├────────────────────────────────────────────┴─────────────┤
│ ──────── │  Timeline                                                │
│ Drop     │  ├─ Commits: [upload] [witness] [reasoning] [...]       │
│ Zone     │  ├─ Locations: ████░░░░████ (with discrepancy markers)  │
│          │  └─ Persons:   ░░████░░░░██                              │
└──────────┴──────────────────────────────────────────────────────────┘
```

---

## 16. 假设与约束（明确写在 README 里）

* 扫描输入形式：多张图片（优先）/视频帧（可选）
* 若没有相机位姿/深度：重建质量下降，但可降级为“对象级 SceneGraph + 粗定位”
* Nano Banana 的出图结果不作为 truth，只作为展示与报告资产
* 推理依赖结构化 SceneGraph，缺失信息将产生 uncertainty

---

如果你愿意，我可以把这份 spec 再进一步“工程化”，补齐两块你们直接开工最需要的东西：

1. **SceneGraph JSON Schema（严格字段定义）** ✅ 已补充（见 5.2 节）
2. **所有 REST/WS 的请求响应示例 + jobs input/output 样例** ✅ 已补充（见 8.1、9.2-9.5 节）

---

## 附录 A：外部依赖与 API 参考

| 服务 | 文档链接 | 用途 |
|------|----------|------|
| Gemini API | https://ai.google.dev/gemini-api/docs | 推理 + 图像生成 |
| Nano Banana | https://ai.google.dev/gemini-api/docs/image-generation | 画像生成 |
| Gemini Thinking | https://ai.google.dev/gemini-api/docs/thinking | 可解释推理 |
| Hunyuan3D-2 | https://github.com/Tencent-Hunyuan/Hunyuan3D-2 | 3D 资产生成 |
| HunyuanWorld-Mirror | https://github.com/Tencent-Hunyuan/HunyuanWorld-Mirror | 场景重建 |
| HunyuanImage-3.0 | https://github.com/Tencent-Hunyuan/HunyuanImage-3.0 | 图像理解 |
| Supabase | https://supabase.com/docs | 数据库 + 存储 + 实时订阅 |

---

*文档版本：v0.2 | 最后更新：2025-09*
