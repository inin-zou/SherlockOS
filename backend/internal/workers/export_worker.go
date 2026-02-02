package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

// ExportWorker handles JobTypeExport jobs
type ExportWorker struct {
	*BaseWorker
	storage clients.StorageClient
}

// NewExportWorker creates a new export worker
func NewExportWorker(database *db.DB, q queue.JobQueue, storage clients.StorageClient) *ExportWorker {
	return &ExportWorker{
		BaseWorker: NewBaseWorker(database, q),
		storage:    storage,
	}
}

// Type returns the job type this worker handles
func (w *ExportWorker) Type() models.JobType {
	return models.JobTypeExport
}

// Process generates an HTML report for the case
func (w *ExportWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input struct {
		Format string `json:"format"`
	}
	if err := json.Unmarshal(job.Input, &input); err != nil {
		return NewFatalError(fmt.Errorf("failed to parse input: %w", err))
	}

	// Update progress
	w.UpdateJobProgress(ctx, job.JobID, 10)

	// Get case data
	caseData, err := w.repo.GetCase(ctx, job.CaseID)
	if err != nil {
		return NewFatalError(fmt.Errorf("failed to get case: %w", err))
	}

	// Get commits (timeline)
	commitPtrs, err := w.repo.GetCommitsByCase(ctx, job.CaseID, 100, nil)
	if err != nil {
		return NewRetryableError(fmt.Errorf("failed to get commits: %w", err))
	}
	// Convert to value slice
	commits := make([]models.Commit, len(commitPtrs))
	for i, c := range commitPtrs {
		commits[i] = *c
	}

	// Update progress
	w.UpdateJobProgress(ctx, job.JobID, 30)

	// Get scene snapshot
	snapshot, err := w.repo.GetSceneSnapshot(ctx, job.CaseID)
	if err != nil {
		// Snapshot might not exist yet, continue with empty
		snapshot = &models.SceneSnapshot{
			CaseID:     job.CaseID,
			Scenegraph: models.NewEmptySceneGraph(),
		}
	}

	// Get suspect profile
	profile, err := w.repo.GetSuspectProfile(ctx, job.CaseID)
	if err != nil {
		// Profile might not exist yet
		profile = nil
	}

	// Update progress
	w.UpdateJobProgress(ctx, job.JobID, 50)

	// Generate HTML report
	reportHTML, err := generateHTMLReport(caseData, commits, snapshot, profile)
	if err != nil {
		return NewRetryableError(fmt.Errorf("failed to generate report: %w", err))
	}

	// Update progress
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Upload report to storage
	storageKey := fmt.Sprintf("cases/%s/reports/report_%s.html", job.CaseID, time.Now().Format("20060102_150405"))
	uploadSucceeded := false

	if w.storage != nil {
		err = w.storage.Upload(ctx, "assets", storageKey, []byte(reportHTML), "text/html")
		if err != nil {
			// Log warning but don't fail - report was generated successfully
			log.Printf("Warning: failed to upload report to storage: %v (report generated locally)", err)
		} else {
			uploadSucceeded = true
		}
	}

	// Mark job as done
	output := map[string]interface{}{
		"format":       "html",
		"generated_at": time.Now().Format(time.RFC3339),
	}
	if uploadSucceeded {
		output["report_asset_key"] = storageKey
	}

	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, output)

	return nil
}

func generateHTMLReport(
	caseData *models.Case,
	commits []models.Commit,
	snapshot *models.SceneSnapshot,
	profile *models.SuspectProfile,
) (string, error) {
	const reportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Case Report: {{.Case.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0a0a0c;
            color: #f0f0f2;
            line-height: 1.6;
            padding: 2rem;
        }
        .container { max-width: 900px; margin: 0 auto; }
        header {
            border-bottom: 1px solid #2a2a32;
            padding-bottom: 1.5rem;
            margin-bottom: 2rem;
        }
        h1 { font-size: 2rem; color: #f0f0f2; }
        h2 { font-size: 1.25rem; color: #a0a0a8; margin: 2rem 0 1rem; border-bottom: 1px solid #1e1e24; padding-bottom: 0.5rem; }
        h3 { font-size: 1rem; color: #8b5cf6; margin: 1rem 0 0.5rem; }
        .meta { color: #606068; font-size: 0.875rem; margin-top: 0.5rem; }
        .section { background: #111114; border: 1px solid #1e1e24; border-radius: 8px; padding: 1.5rem; margin-bottom: 1rem; }
        .badge { display: inline-block; padding: 0.25rem 0.75rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 500; }
        .badge-blue { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
        .badge-green { background: rgba(34, 197, 94, 0.1); color: #22c55e; }
        .badge-amber { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
        .badge-purple { background: rgba(139, 92, 246, 0.1); color: #8b5cf6; }
        .timeline { list-style: none; }
        .timeline li { position: relative; padding: 1rem 0 1rem 2rem; border-left: 2px solid #2a2a32; }
        .timeline li::before { content: ''; position: absolute; left: -5px; top: 1.25rem; width: 8px; height: 8px; border-radius: 50%; background: #3b82f6; }
        .grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; }
        .attribute { display: flex; justify-content: space-between; padding: 0.5rem 0; border-bottom: 1px solid #1e1e24; }
        .attribute-label { color: #606068; }
        .attribute-value { color: #f0f0f2; }
        .evidence-card { background: #1f1f24; padding: 1rem; border-radius: 6px; margin-bottom: 0.75rem; }
        .evidence-title { font-weight: 500; margin-bottom: 0.5rem; }
        .evidence-desc { color: #a0a0a8; font-size: 0.875rem; }
        footer { margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid #2a2a32; text-align: center; color: #606068; font-size: 0.75rem; }
        .logo { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; justify-content: center; }
        .logo-icon { width: 24px; height: 24px; background: linear-gradient(135deg, #6366f1, #8b5cf6); border-radius: 6px; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>{{.Case.Title}}</h1>
            <p class="meta">Case ID: {{.Case.ID}} | Generated: {{.GeneratedAt}}</p>
            {{if .Case.Description}}<p style="margin-top: 0.5rem; color: #a0a0a8;">{{.Case.Description}}</p>{{end}}
        </header>

        <h2>Timeline ({{len .Commits}} events)</h2>
        <ul class="timeline">
            {{range .Commits}}
            <li>
                <div style="display: flex; justify-content: space-between; align-items: center;">
                    <span class="badge badge-blue">{{.Type}}</span>
                    <span class="meta">{{.CreatedAt.Format "Jan 2, 2006 3:04 PM"}}</span>
                </div>
                <p style="margin-top: 0.5rem;">{{.Summary}}</p>
            </li>
            {{end}}
        </ul>

        {{if .HasProfile}}
        <h2>Suspect Profile</h2>
        <div class="section">
            <p style="color: #a0a0a8;">Profile attributes on file.</p>
        </div>
        {{end}}

        <h2>Evidence ({{len .Evidence}} items)</h2>
        {{range .Evidence}}
        <div class="evidence-card">
            <div class="evidence-title">{{.Title}}</div>
            <div class="evidence-desc">{{.Description}}</div>
            <div style="margin-top: 0.5rem;">
                <span class="badge badge-{{if ge .Confidence 0.8}}green{{else if ge .Confidence 0.5}}amber{{else}}blue{{end}}">{{printf "%.0f%%" (mul .Confidence 100)}} confidence</span>
            </div>
        </div>
        {{end}}

        <h2>Scene Summary</h2>
        <div class="section">
            <p style="color: #a0a0a8;">Objects detected: {{len .Objects}}</p>
            {{range .Objects}}
            <div style="padding: 0.5rem 0; border-bottom: 1px solid #1e1e24;">
                <span class="badge badge-blue">{{.Type}}</span>
                <span style="margin-left: 0.5rem;">{{.Label}}</span>
            </div>
            {{end}}
        </div>

        <footer>
            <div class="logo">
                <div class="logo-icon"></div>
                <span>SherlockOS</span>
            </div>
            <p>This report is for investigative purposes only. AI-generated hypotheses should be verified with physical evidence.</p>
            <p style="margin-top: 0.25rem;">Generated by SherlockOS v0.1 | {{.GeneratedAt}}</p>
        </footer>
    </div>
</body>
</html>`

	// Create template functions
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Build template data
	var evidence []models.EvidenceCard
	var objects []models.SceneObject
	if snapshot != nil && snapshot.Scenegraph != nil {
		evidence = snapshot.Scenegraph.Evidence
		objects = snapshot.Scenegraph.Objects
	}

	data := struct {
		Case        *models.Case
		Commits     []models.Commit
		HasProfile  bool
		Evidence    []models.EvidenceCard
		Objects     []models.SceneObject
		GeneratedAt string
	}{
		Case:        caseData,
		Commits:     commits,
		HasProfile:  profile != nil,
		Evidence:    evidence,
		Objects:     objects,
		GeneratedAt: time.Now().Format("January 2, 2006 3:04 PM"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
