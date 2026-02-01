# SherlockOS — Project Demo & Pitch Script

## 1. Introduction: The "Single Source of Truth" Problem
*“In every investigation, there are two versions of the truth: what the digital logs say, and what people remember. The problem is they rarely speak the same language.”*

Welcome to **SherlockOS**. We aren't just building a crime scene viewer; we are building a **Spatio-Temporal Deduction Engine**. SherlockOS takes messy, multimodal evidence and reconciles it into a single, structured 3D world model.

---

## 2. Key Architecture: The Reliability Hierarchy
Our core innovation is the **Reliability Hierarchy (Tiers 0-3)**. 
- **Tier 0 (Environment)**: The unmoving walls and floor plans.
- **Tier 1 (Ground Truth)**: Deniable video records (CCTV).
- **Tier 2 (Digital Logs)**: IoT and sensor data.
- **Tier 3 (Testimonials)**: The fuzzy, human stories.

SherlockOS uses **HunyuanWorld** to reconstruct the static environment and **Gemini 2.5 Flash** to find the "Temporal Paradoxes"—the points where a human story breaks the laws of physics or contradicts a digital log.

---

## 3. Demo Walkthrough Script

### **Step 1: Ingestion (Feeding the Machine)**
*“We begin by feeding SherlockOS every scrap of data. Notice how the system automatically categorizes each file—CCTV becomes a Tier 1 Hard Anchor, while a phone recording of a witness is flagged as Tier 3 Subjective Evidence.”*

### **Step 2: Model (Building the World)**
*“Then, we build the stage. Using **HunyuanWorld**, SherlockOS reconstructs the physical room with **Proxy Geometry**. These blue boxes represent furniture that cannot be passed through or seen through. We now have a mathematically accurate playground for our AI.”*

### **Step 3: Deduction (Finding the Discrepancies)**
*“This is where it gets powerful. We ask Gemini to deduce the **Witness Motion Path**. Based on the testimony, the AI plots where the witness was moving. It then simulates their **Perspective** (POV) at every second. If the witness says they saw a gun, but the Proxy Geometry proves the kitchen island was in the way, SherlockOS flags a **Visual Discrepancy**.”*

### **Step 4: Simulation (Replaying the Truth)**
*“Finally, we run the **Simulation**. SherlockOS combines the digital logs, the video anchors, and the 'validated' parts of the testimony to recreate the most probable version of events in 4D. We don't just see a photo; we see the logic of the crime.”*

---

## 4. Technical Vision (The Google Ecosystem)
Currently, SherlockOS leverages the **Hunyuan World Model** for its high accessibility and open-source flexibility in reconstruction. 

However, we’ve built the system to be modular. We are particularly excited about **Google’s D4RT (Dynamic 4D Reconstruction & Tracking)**. While not currently open-sourced for this demo, SherlockOS’s architecture is designed so that D4RT can be swapped in as the primary engine. This would upgrade our "Ground Truth" layer to full 4D dynamic tracking, making SherlockOS the most advanced forensic tool in the Google ecosystem.

---

## 5. Conclusion
*“SherlockOS doesn't tell you who did it. It shows you what is **impossible**, so only the **truth** remains.”*
