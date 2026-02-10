# SherlockOS Demo Data

This folder contains pre-generated assets for demonstrating SherlockOS capabilities.

## Scenario: Tech Corp Office Break-in

A night-time break-in at Tech Corp headquarters. The suspect:
1. Forced entry through a 3rd floor window
2. Searched the executive desk and filing cabinets
3. Dropped a glove while searching
4. Exited through the emergency door

## Contents

### `/images/` - Crime Scene Photos (Generated with Nano Banana)

| File | Description |
|------|-------------|
| `01_scene_overview.png` | Wide shot of ransacked office |
| `02_entry_point.png` | Broken window (entry point) |
| `03_footprints.png` | Muddy boot prints on carpet |
| `04_desk_area.png` | Ransacked desk with dropped glove |
| `05_exit_door.png` | Emergency exit door |
| `06_evidence_closeup.png` | Collected evidence items |

### `/data/` - Demo Data Files

| File | Description |
|------|-------------|
| `case_info.json` | Case metadata and timeline |
| `witness_statements.json` | 4 witness statements with credibility scores |
| `scene_images.json` | Image metadata and evidence markers |

## How to Use in Demo

### 1. Create a Case
```bash
curl -X POST http://localhost:8080/v1/cases \
  -H "Content-Type: application/json" \
  -d @data/case_info.json
```

### 2. Upload Scene Images
Upload all images from `/images/` folder to the case.

### 3. Add Witness Statements
```bash
curl -X POST http://localhost:8080/v1/cases/{case_id}/commits \
  -H "Content-Type: application/json" \
  -d '{
    "type": "witness_statement",
    "summary": "Security guard statement",
    "payload": <statement_from_json>
  }'
```

### 4. Trigger Profile Extraction
Submit a profile job with all witness statements to extract suspect attributes.

### 5. Run Trajectory Reasoning
Once scene is reconstructed, run reasoning to generate movement hypotheses.

## Expected Demo Flow

1. **Scene Upload** → Watch 3D scene build progressively
2. **Evidence Detection** → AI identifies key items
3. **Witness Input** → Add statements, watch profile evolve
4. **Portrait Generation** → Generate suspect image from merged attributes
5. **Trajectory Analysis** → See movement path hypotheses
6. **Replay Simulation** → Video of most likely trajectory

## Suspect Profile (Expected from Statements)

Based on the witness statements, the system should extract:

| Attribute | Value | Confidence |
|-----------|-------|------------|
| Height | ~6'0" (tall) | 0.85 |
| Build | Slim | 0.75 |
| Age | 30s | 0.6 |
| Hair | Short, dark | 0.6 |
| Facial Hair | Beard | 0.6 |
| Clothing | Dark hoodie, jeans, cap | 0.7 |
| Notable | Limping (right leg) | 0.9 |
| Vehicle | White sedan | 0.6 |

Note: Conflicting observations (hoodie vs cap, tall vs medium height) should be flagged by the system.
