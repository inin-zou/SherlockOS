# SherlockOS — Technical Specification (v0.2)

## 1. 项目概述

SherlockOS 是一个基于“世界模型（World Model）+ 证据可靠性层级（Reliability Hierarchy）”的侦查辅助系统。它通过将不同来源的证据（Tier 0-3）映射到统一的空间与时间线上，利用 AI 发现时空矛盾（Spatial/Temporal Paradox），还原案件逻辑真相。

### 1.1 核心原则
1. **可靠性层级（Reliability Hierarchy）**：根据证据的客观性自动分配权重，从核心物理边界到主观证词。
2. **空间锚定（Spatial Anchoring）**：所有证据在 3D/2.5D 空间中拥有坐标或观察角度。
3. **时空辩证（Spatio-Temporal Deduction）**：利用硬性锚点（Hard Anchors）修正模糊证词（Soft Events）。

---

## 2. 证据可靠性层级 (Reliability Hierarchy)

| 级别 | 名称 | 定义 | 典型输入 | 逻辑权重 |
| :--- | :--- | :--- | :--- | :--- |
| **Tier 0** | **Environment** | 物理基础设施 / 不可穿透边界 | 平面图、静态扫描、HunyuanWorld 重建 | **100% / 物理底座** |
| **Tier 1** | **Ground Truth** | 原始视觉/录音记录 | CCTV、记录仪、关键帧视频 | **高 / 时空硬锚点** |
| **Tier 2** | **Deterministic Logs** | 数字化触发记录 | 智能锁日志、Wi-Fi 连接、传感器 | **中 / 真实但存歧义** |
| **Tier 3** | **Testimonials** | 主观描述 | 目击者证词、嫌疑人辩解 | **低 / 主观概率** |

---

## 3. 技术栈与模型能力 (Hunyuan Ecosystem focus)

### 3.1 核心 AI 模型
*   **HunyuanWorld-Mirror / 1.0**: 核心场景重建引擎。负责从多视角图片/视频帧还原 **Tier 0 环境**。
*   **Hunyuan3D-2.1**: 负责从发现的证据项（Weapon, Footprint）生成高质量 3D 资产。
*   **Gemini 2.5 Flash (Thinking Mode)**: 核心逻辑推理大脑。
    *   **空间分析**：进行视线验证与物理通行路径判断。
    *   **时空对齐**：将 Tier 3 证词中的模糊时间锚定到 Tier 1/2 的硬时间戳。
*   **Nano Banana (Gemini Image Gen)**: 生成高保真嫌疑人画像。

### 3.2 2.5D 与 Proxy Geometry（代理几何体）
*   **Proxy Geometry**: 利用 HunyuanWorld 生成的边界，系统将物体转化为简单的几何体（Box/Cylinder）。Gemini 在这些简化的数学结构上运行空间逻辑，速度与准确率远高于处理复杂网格。
*   **2.5D Overlay**: 在原始侦查照片上叠加深度值，支持侦查员进行虚拟测距。

---

## 4. 侦查辅助核心逻辑

### 4.1 Chronological Analysis (时空映射)
系统利用 **Hard Anchor Mapping** 机制。例如：
1. **硬锚点**: CCTV 显示嫌疑人 22:05 进入。
2. **模糊证词**: 目击者称“开门后大约几分钟听到响动”。
3. **演绎**: 系统将“响动”映射为 22:08±60s，并由 Gemini 检查该时间内其他传感器的并发记录。

### 4.2 Discrepancy Detection (矛盾发现)
*   **Sightline Paradox (视线矛盾)**：AI 计算证明目击者位置与宣称看到的物体间存在墙壁（Tier 0）。
*   **Temporal Paradox (时间悖论)**：嫌疑人称在此处，但 Tier 2 日志显示其设备在彼处。

---

## 5. 工作流 (The Sherlock Pipeline)

1.  **Ingestion (证据摄入)**:
    *   摄入多模态证据并根据 Tier (0-3) 自动分配初始权重。
2.  **Model (世界建模)**:
    *   **HunyuanWorld** 构建静态物理环境与 **Proxy Geometry（代理几何体）**。
    *   建立空间基准坐标系，确定“硬性不可穿越”边界。
3.  **Deduction (推演与矛盾探测)**:
    *   **Witness Path Calculation**: Gemini 提取证词，推算目击者的运动路径（Motion Path）。
    *   **Perspective Validation**: 在 3D 空间中模拟目击者的视点（POV），验证宣称的观察是否受物理遮挡影响。
    *   **Discrepancy Identification**: 对比 Tier 3 (证词) 与 Tier 0/1/2，高亮显示不合逻辑的“时空矛盾”节点。
4.  **Simulation (真实还原仿真)**:
    *   将所有经过验证的“硬事实”与校准后的“软证据”结合，进行 4D 动态全景仿真。
    *   生成“最优拟合”的案件经过复现。

---

## 6. 技术演进 (Future-proofing)
虽然当前使用 **HunyuanWorld** 作为重建引擎，但系统架构完全兼容 Google **D4RT (Dynamic 4D Reconstruction & Tracking)**。未来若 D4RT 开放接入，通过简单的 Worker 替换，即可实现更极致的 4D 动态追踪与视频-空间无缝映射，深度融入 Google 侦查生态。
