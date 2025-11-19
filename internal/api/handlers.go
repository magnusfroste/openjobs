package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"openjobs/internal/scheduler"
	"openjobs/pkg/models"
	"openjobs/pkg/storage"

	"github.com/google/uuid"
)

// Server holds the HTTP server dependencies
type Server struct {
	jobStore  *storage.JobStore
	scheduler *scheduler.Scheduler
}

// NewServer creates a new server instance
func NewServer(jobStore *storage.JobStore, scheduler *scheduler.Scheduler) *Server {
	return &Server{
		jobStore:  jobStore,
		scheduler: scheduler,
	}
}

// DashboardHandler serves the professional analytics dashboard
// DEPRECATED: Dashboard moved to OpenJobs_Web (https://openjobs-web.vercel.app)
// This handler is kept for backward compatibility but is no longer used
// TODO: Remove in future version after confirming no dependencies
func (s *Server) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenJobs Dashboard</title>
    <style>
        :root {
            --bg-gradient-start: #667eea;
            --bg-gradient-end: #764ba2;
            --card-bg: white;
            --text-primary: #2c3e50;
            --text-secondary: #7f8c8d;
            --border-color: #e1e8ed;
            --table-header-bg: #f8f9fa;
            --table-hover-bg: #f8f9fa;
            --shadow: rgba(0, 0, 0, 0.1);
        }
        
        [data-theme="dark"] {
            --bg-gradient-start: #1a1a2e;
            --bg-gradient-end: #16213e;
            --card-bg: #0f3460;
            --text-primary: #e8e8e8;
            --text-secondary: #a8a8a8;
            --border-color: #1a4d6d;
            --table-header-bg: #16213e;
            --table-hover-bg: #1a4d6d;
            --shadow: rgba(0, 0, 0, 0.3);
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, var(--bg-gradient-start) 0%, var(--bg-gradient-end) 100%);
            min-height: 100vh;
            padding: 2rem;
            transition: background 0.3s ease;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 2rem;
        }
        .header {
            background: var(--card-bg);
            padding: 2rem;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.08);
            margin-bottom: 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            font-size: 1.75rem;
            color: var(--text-primary);
            margin-bottom: 0.25rem;
        }
        .subtitle {
            color: var(--text-secondary);
            font-size: 0.9rem;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        .stat-card {
            background: var(--card-bg);
            padding: 1.5rem;
            border-radius: 12px;
            box-shadow: 0 2px 8px var(--shadow);
            transition: background 0.3s ease;
        }
        .stat-label {
            color: var(--text-secondary);
            font-size: 0.85rem;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 0.5rem;
        }
        .stat-value {
            font-size: 2rem;
            font-weight: 700;
            color: var(--text-primary);
        }
        .section {
            background: var(--card-bg);
            padding: 2rem;
            border-radius: 12px;
            box-shadow: 0 2px 8px var(--shadow);
            margin-bottom: 2rem;
            transition: background 0.3s ease;
        }
        .section-title {
            font-size: 1.25rem;
            color: var(--text-primary);
            margin-bottom: 1.5rem;
            font-weight: 600;
        }
        .plugin-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1rem;
        }
        .plugin-card {
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 1.25rem;
            transition: border-color 0.3s ease;
            margin-bottom: 1.25rem;
        }
        .plugin-name {
            font-weight: 600;
            color: var(--text-primary);
        }
        .status-badge {
            padding: 0.5rem 0.75rem;
            border-radius: 12px;
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
        }
        .status-healthy { background: #d4edda; color: #155724; }
        .status-warning { background: #fff3cd; color: #856404; }
        .plugin-stats {
            display: flex;
            gap: 1rem;
            font-size: 0.85rem;
            color: var(--text-secondary);
        }
        .action-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.9rem;
            transition: background 0.2s;
        }
        .action-btn:hover {
            background: #2980b9;
        }
        .action-btn:disabled {
            background: #95a5a6;
            cursor: not-allowed;
        }
        .loading { color: #95a5a6; }
        .sync-log-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.9rem;
        }
        .sync-log-table th {
            text-align: left;
            padding: 0.75rem;
            background: var(--table-header-bg);
            border-bottom: 2px solid var(--border-color);
            font-weight: 600;
            color: var(--text-primary);
        }
        .sync-log-table td {
            padding: 0.75rem;
            border-bottom: 1px solid var(--border-color);
            color: var(--text-primary);
        }
        .sync-log-table tr:hover {
            background: var(--table-hover-bg);
        }
        .efficiency-badge {
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.8rem;
            font-weight: 600;
        }
        .efficiency-high { background: #d4edda; color: #155724; }
        .efficiency-medium { background: #fff3cd; color: #856404; }
        .efficiency-low { background: #f8d7da; color: #721c24; }
        @media (max-width: 768px) {
            .container { padding: 1rem; }
            .header { flex-direction: column; align-items: flex-start; }
            .stats-grid { grid-template-columns: 1fr; }
            .sync-log-table { font-size: 0.8rem; }
            .sync-log-table th, .sync-log-table td { padding: 0.5rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div>
                <h1>OpenJobs Dashboard</h1>
                <div class="subtitle">Microservices Job Aggregation Platform</div>
            </div>
            <div style="display: flex; gap: 1rem; align-items: center;">
                <button class="action-btn" onclick="toggleTheme()" style="background: #95a5a6;" title="Toggle dark mode">
                    <span id="theme-icon">🌙</span>
                </button>
                <button class="action-btn" onclick="triggerSync()">Sync All Plugins</button>
            </div>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-label">Total Jobs</div>
                <div class="stat-value" id="total-jobs"><span class="loading">--</span></div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Active Plugins</div>
                <div class="stat-value" id="plugins"><span class="loading">--</span></div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Remote Jobs</div>
                <div class="stat-value" id="remote"><span class="loading">--</span></div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Last Sync</div>
                <div class="stat-value" style="font-size: 1.2rem;" id="last-sync"><span class="loading">--</span></div>
            </div>
        </div>

        <div class="section">
            <div class="section-title">Plugin Status</div>
            <div class="plugin-grid" id="plugin-status">
                <div class="loading">Loading plugins...</div>
            </div>
        </div>

        <div class="section">
            <div class="section-title">Sync History</div>
            <div id="sync-log">
                <div class="loading">Loading sync history...</div>
            </div>
        </div>

        <div class="section">
            <div class="section-title">API Endpoints</div>
            <div style="font-family: monospace; font-size: 0.9rem; line-height: 2;">
                <div><strong>GET</strong> /health - Health check</div>
                <div><strong>GET</strong> /jobs - List all jobs</div>
                <div><strong>GET</strong> /jobs/{id} - Get specific job</div>
                <div><strong>POST</strong> /sync/manual - Trigger manual sync</div>
                <div><strong>GET</strong> /analytics - Get analytics data</div>
            </div>
        </div>

        <!-- Version footer -->
        <div style="text-align: center; margin-top: 3rem; padding: 1.5rem; color: var(--text-secondary); font-size: 0.75rem; opacity: 0.6;">
            <div id="version-info">OpenJobs v<span id="app-version">-</span></div>
            <div id="build-time" style="font-size: 0.7rem; margin-top: 0.25rem;"></div>
        </div>
    </div>

    <script>
        async function loadDashboard() {
            try {
                const [analyticsRes, healthRes] = await Promise.all([
                    fetch('/analytics'),
                    fetch('/health')
                ]);
                
                const analyticsData = await analyticsRes.json();
                const healthData = await healthRes.json();

                // Update version info
                if (healthData.data?.version) {
                    document.getElementById('app-version').textContent = healthData.data.version;
                    if (healthData.data.build_time && healthData.data.build_time !== 'unknown') {
                        document.getElementById('build-time').textContent = 'Built: ' + healthData.data.build_time;
                    }
                }

                // Update stats from analytics summary
                const summary = analyticsData.data?.summary;
                if (summary) {
                    document.getElementById('total-jobs').textContent = summary.total_jobs || '0';
                    document.getElementById('plugins').textContent = summary.sources_count || '4';
                    document.getElementById('remote').textContent = summary.remote_percentage + '%' || '0%';
                    
                    // Show last sync from analytics
                    if (summary.last_sync) {
                        const lastSync = new Date(summary.last_sync);
                        const hours = String(lastSync.getHours()).padStart(2, '0');
                        const minutes = String(lastSync.getMinutes()).padStart(2, '0');
                        const day = String(lastSync.getDate()).padStart(2, '0');
                        const month = String(lastSync.getMonth() + 1).padStart(2, '0');
                        document.getElementById('last-sync').textContent = hours + ':' + minutes + ' ' + day + '/' + month;
                    } else {
                        document.getElementById('last-sync').textContent = 'Never';
                    }
                } else {
                    // Fallback values
                    document.getElementById('total-jobs').textContent = '0';
                    document.getElementById('plugins').textContent = '4';
                    document.getElementById('remote').textContent = '0%';
                    document.getElementById('last-sync').textContent = 'Never';
                }

                // Fetch and update plugin status from API
                const pluginRes = await fetch('/plugins/status');
                const pluginData = await pluginRes.json();

                let pluginHtml = '';
                if (pluginData.success && pluginData.data) {
                    const activePlugins = pluginData.data.filter(p => p.status === 'healthy').length;
                    document.getElementById('plugins').textContent = activePlugins;

                    for (const plugin of pluginData.data) {
                        pluginHtml += '<div class="plugin-card">' +
                            '<div class="plugin-header">' +
                            '<div class="plugin-name">' + plugin.name + '</div>' +
                            '<span class="status-badge status-' + plugin.status + '">&bull; ' + plugin.status + '</span>' +
                            '</div>' +
                            '<div class="plugin-stats">' +
                            '<div>Port: ' + plugin.port + '</div>' +
                            '<div>Jobs: ' + plugin.jobs + '</div>' +
                            '</div>' +
                            '</div>';
                    }
                } else {
                    // Fallback to hardcoded data
                    const plugins = [
                        { name: 'Arbetsförmedlingen', port: 8081, jobs: 0, status: 'unknown' },
                        { name: 'EURES', port: 8082, jobs: 0, status: 'unknown' },
                        { name: 'Remotive', port: 8083, jobs: 0, status: 'unknown' },
                        { name: 'RemoteOK', port: 8084, jobs: 0, status: 'unknown' }
                    ];
                    for (const plugin of plugins) {
                        pluginHtml += '<div class="plugin-card">' +
                            '<div class="plugin-header">' +
                            '<div class="plugin-name">' + plugin.name + '</div>' +
                            '<span class="status-badge status-' + plugin.status + '">&bull; ' + plugin.status + '</span>' +
                            '</div>' +
                            '<div class="plugin-stats">' +
                            '<div>Port: ' + plugin.port + '</div>' +
                            '<div>Jobs: ' + plugin.jobs + '</div>' +
                            '</div>' +
                            '</div>';
                    }
                }
                document.getElementById('plugin-status').innerHTML = pluginHtml;

                // Fetch real sync logs from API
                const logsRes = await fetch('/sync/logs');
                const logsData = await logsRes.json();

                if (logsData.success && logsData.data && logsData.data.length > 0) {
                    let logHtml = '<table class="sync-log-table"><thead><tr>' +
                        '<th>Plugin</th><th>Time</th><th>Fetched</th><th>Inserted</th><th>Duplicates</th><th>Efficiency</th>' +
                        '</tr></thead><tbody>';
                    
                    logsData.data.forEach(log => {
                        const efficiency = log.jobs_fetched > 0 ? Math.round((log.jobs_inserted / log.jobs_fetched) * 100) : 0;
                        const efficiencyClass = efficiency > 80 ? 'efficiency-high' : efficiency > 50 ? 'efficiency-medium' : 'efficiency-low';
                        const timeAgo = getTimeAgo(new Date(log.started_at));
                        const pluginName = log.connector_name.charAt(0).toUpperCase() + log.connector_name.slice(1);
                        
                        logHtml += '<tr>' +
                            '<td><strong>' + pluginName + '</strong></td>' +
                            '<td>' + timeAgo + '</td>' +
                            '<td>' + log.jobs_fetched + '</td>' +
                            '<td><strong>' + log.jobs_inserted + '</strong></td>' +
                            '<td>' + log.jobs_duplicates + '</td>' +
                            '<td><span class="efficiency-badge ' + efficiencyClass + '">' + efficiency + '%</span></td>' +
                            '</tr>';
                    });
                    
                    logHtml += '</tbody></table>';
                    document.getElementById('sync-log').innerHTML = logHtml;
                } else {
                    document.getElementById('sync-log').innerHTML = '<div style="color: #7f8c8d; text-align: center; padding: 2rem;">No sync history yet. Run a sync to see data.</div>';
                }

            } catch (error) {
                console.error('Error loading dashboard:', error);
                document.getElementById('total-jobs').textContent = 'Error';
                document.getElementById('plugins').textContent = 'Error';
                document.getElementById('remote').textContent = 'Error';
                document.getElementById('last-sync').textContent = 'Error';
            }
        }

        async function triggerSync() {
            const btn = event.target;
            btn.disabled = true;
            btn.textContent = 'Syncing...';
            
            try {
                await fetch('/sync/manual', { method: 'POST' });
                btn.textContent = 'Sync Complete!';
                setTimeout(() => {
                    btn.disabled = false;
                    btn.textContent = 'Sync All Plugins';
                    loadDashboard();
                }, 2000);
            } catch (error) {
                btn.textContent = 'Sync Failed';
                setTimeout(() => {
                    btn.disabled = false;
                    btn.textContent = 'Sync All Plugins';
                }, 2000);
            }
        }

        function getTimeAgo(date) {
            // Format: HH:MM DD/MM
            const d = new Date(date);
            const hours = String(d.getHours()).padStart(2, '0');
            const minutes = String(d.getMinutes()).padStart(2, '0');
            const day = String(d.getDate()).padStart(2, '0');
            const month = String(d.getMonth() + 1).padStart(2, '0');
            return hours + ':' + minutes + ' ' + day + '/' + month;
        }

        function toggleTheme() {
            const html = document.documentElement;
            const currentTheme = html.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            const icon = document.getElementById('theme-icon');
            
            html.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            icon.textContent = newTheme === 'dark' ? '☀️' : '🌙';
        }

        function initTheme() {
            const savedTheme = localStorage.getItem('theme') || 'light';
            const icon = document.getElementById('theme-icon');
            document.documentElement.setAttribute('data-theme', savedTheme);
            icon.textContent = savedTheme === 'dark' ? '☀️' : '🌙';
        }

        document.addEventListener('DOMContentLoaded', () => {
            initTheme();
            loadDashboard();
        });
        setInterval(loadDashboard, 30000); // Refresh every 30 seconds
    </script>
</body>
</html>`
	w.Write([]byte(html))
}

// DashboardHandlerAlternative serves the analytics dashboard as a fallback
// DEPRECATED: Dashboard moved to OpenJobs_Web (https://openjobs-web.vercel.app)
// This handler is kept for backward compatibility but is no longer used
// TODO: Remove in future version after confirming no dependencies
func (s *Server) DashboardHandlerAlternative(w http.ResponseWriter, r *http.Request) {
	s.DashboardHandler(w, r)
}

// AnalyticsHandler handles GET /analytics - Analytics dashboard
func (s *Server) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get real data from database
	totalJobs, _ := s.jobStore.GetTotalJobCount()
	remoteJobs, _ := s.jobStore.GetRemoteJobCount()
	remotePercentage := 0
	if totalJobs > 0 {
		remotePercentage = (remoteJobs * 100) / totalJobs
	}

	// Get analytics by source from materialized view
	sourceAnalytics, err := s.jobStore.GetAnalyticsBySource()
	if err != nil {
		fmt.Printf("Error getting analytics by source: %v\n", err)
		sourceAnalytics = []map[string]interface{}{} // Empty array on error
	}

	// Build sources array from analytics
	sources := make([]map[string]interface{}, 0, len(sourceAnalytics))
	sourcesCount := len(sourceAnalytics)
	totalCountries := 0

	for _, analytics := range sourceAnalytics {
		source := map[string]interface{}{
			"source":      analytics["source"],
			"total_jobs":  analytics["total_jobs"],
			"active_jobs": analytics["active_jobs"],
			"remote_jobs": analytics["remote_jobs"],
		}

		// Add optional fields if they exist
		if countries, ok := analytics["countries_covered"]; ok {
			source["countries_covered"] = countries
			if c, ok := countries.(float64); ok {
				totalCountries += int(c)
			}
		}

		if avgMin, ok := analytics["avg_min_salary"]; ok && avgMin != nil {
			source["avg_min_salary"] = avgMin
		}

		if avgMax, ok := analytics["avg_max_salary"]; ok && avgMax != nil {
			source["avg_max_salary"] = avgMax
		}

		if latestSync, ok := analytics["latest_sync"]; ok && latestSync != nil {
			source["latest_sync"] = latestSync
		}

		sources = append(sources, source)
	}

	// Get last sync time from sync logs
	lastSyncTime := time.Now().Format(time.RFC3339)
	logs, err := s.jobStore.GetRecentSyncLogs(1)
	if err == nil && len(logs) > 0 {
		lastSyncTime = logs[0].StartedAt.Format(time.RFC3339)
	}

	// Get recent activity from sync logs
	activity := make([]map[string]interface{}, 0)
	recentLogs, err := s.jobStore.GetRecentSyncLogs(10)
	if err == nil {
		for _, log := range recentLogs {
			event := "sync_completed"
			if log.Status != "success" {
				event = "sync_failed"
			}

			details := fmt.Sprintf("Fetched %d jobs, inserted %d new jobs",
				log.JobsFetched, log.JobsInserted)
			if log.ErrorMessage != "" {
				details = log.ErrorMessage
			}

			activity = append(activity, map[string]interface{}{
				"timestamp": log.StartedAt.Format(time.RFC3339),
				"event":     event,
				"source":    log.ConnectorName,
				"details":   details,
			})
		}
	}

	// Return real analytics data
	response := models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"summary": map[string]interface{}{
				"total_jobs":        totalJobs,
				"sources_count":     sourcesCount,
				"countries_covered": totalCountries,
				"remote_percentage": remotePercentage,
				"last_sync":         lastSyncTime,
			},
			"sources":  sources,
			"activity": activity,
		},
		Message: "Analytics data retrieved successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// ValidateAPIKey middleware validates the X-API-Key header against companies table
func (s *Server) ValidateAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, `{"success": false, "message": "Missing X-API-Key header"}`, http.StatusUnauthorized)
			return
		}

		// Validate API key against Supabase companies table
		supabaseURL := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

		url := fmt.Sprintf("%s/rest/v1/companies?api_key=eq.%s&select=id,name,email", supabaseURL, apiKey)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Internal error"}`, http.StatusInternalServerError)
			return
		}

		req.Header.Set("apikey", supabaseKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", supabaseKey))

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Failed to validate API key"}`, http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var companies []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&companies); err != nil || len(companies) == 0 {
			http.Error(w, `{"success": false, "message": "Invalid API key"}`, http.StatusUnauthorized)
			return
		}

		// API key is valid, proceed to next handler
		next(w, r)
	}
}

// CreateJob handles POST /jobs with API key validation
func (s *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var job models.JobPost
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if job.Title == "" {
		http.Error(w, `{"success": false, "message": "Title is required"}`, http.StatusBadRequest)
		return
	}
	if job.Company == "" {
		http.Error(w, `{"success": false, "message": "Company is required"}`, http.StatusBadRequest)
		return
	}
	if job.Description == "" {
		http.Error(w, `{"success": false, "message": "Description is required"}`, http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if job.ID == "" {
		job.ID = fmt.Sprintf("web-%s", uuid.New().String())
	}

	// Set timestamps
	now := time.Now()

	// posted_date: When the employer posted the job (user can specify, defaults to now)
	// This is different from created_at (when OpenJobs created the record)
	if job.PostedDate.IsZero() {
		job.PostedDate = now // Default to now if not specified
	}

	// expires_date: User can specify, or leave empty (no expiration)

	// Set default values
	job.IsActive = true // Jobs posted via API are active by default

	// Set source if not provided
	if job.Source == "" {
		job.Source = "openjobs-api" // Default for direct API posts
	}

	// Soft validation: log warning for unusual sources (but allow them)
	knownSources := map[string]bool{
		"arbetsformedlingen": true, "remoteok": true, "remotive": true,
		"adzuna": true, "indeed": true, "jooble": true,
		"openjobs-web": true, "openjobs-api": true,
		"greenhouse-api": true, "workable-api": true, "lever-api": true,
		"custom-integration": true,
	}
	if !knownSources[job.Source] {
		fmt.Printf("⚠️  Unusual source detected: '%s' (allowed but not in standard list)\n", job.Source)
	}

	// Initialize fields map if nil and add metadata
	if job.Fields == nil {
		job.Fields = make(map[string]interface{})
	}

	job.Fields["api_posted"] = true
	job.Fields["posted_at"] = now.Format(time.RFC3339)

	if err := s.jobStore.CreateJob(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Failed to create job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    job,
		Message: "Job created successfully",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetAllJobs handles GET /jobs
func (s *Server) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	limit := 20 // default
	offset := 0
	createdAfter := r.URL.Query().Get("created_after")
	company := r.URL.Query().Get("company") // Filter by company name

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Support incremental sync via created_after parameter
	var jobs []*models.JobPost
	var err error

	// Filter by company if specified
	if company != "" {
		jobs, err = s.jobStore.GetJobsByCompany(company, limit, offset)
	} else if createdAfter != "" {
		jobs, err = s.jobStore.GetJobsAfter(createdAfter, limit, offset)
	} else {
		jobs, err = s.jobStore.GetAllJobs(limit, offset)
	}

	if err != nil {
		http.Error(w, `{"success": false, "message": "Failed to retrieve jobs"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    jobs,
	}

	json.NewEncoder(w).Encode(response)
}

// SyncJobs handles POST /sync/manual - Manual job synchronization
func (s *Server) SyncJobs(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🔧 Manual sync requested: %s %s\n", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, `{"success": false, "message": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Run manual sync
	err := s.scheduler.RunManualSync()
	if err != nil {
		response := models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Sync failed: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Job synchronization completed successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetJobByID handles GET /jobs/{id}
func (s *Server) GetJobByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/jobs/"):] // Extract ID from path
	if id == "" {
		http.Error(w, `{"success": false, "message": "Job ID required"}`, http.StatusBadRequest)
		return
	}

	job, err := s.jobStore.GetJob(id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, `{"success": false, "message": "Job not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"success": false, "message": "Failed to retrieve job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    job,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateJob handles PUT /jobs/{id}
func (s *Server) UpdateJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/jobs/"):] // Extract ID from path
	if id == "" {
		http.Error(w, `{"success": false, "message": "Job ID required"}`, http.StatusBadRequest)
		return
	}

	var job models.JobPost
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	job.ID = id // Ensure ID matches path

	if err := s.jobStore.UpdateJob(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Failed to update job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    job,
		Message: "Job updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteJob handles DELETE /jobs/{id}
func (s *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/jobs/"):] // Extract ID from path
	if id == "" {
		http.Error(w, `{"success": false, "message": "Job ID required"}`, http.StatusBadRequest)
		return
	}

	if err := s.jobStore.DeleteJob(id); err != nil {
		http.Error(w, `{"success": false, "message": "Failed to delete job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Job deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// SyncLogsHandler handles GET /sync/logs - Get recent sync logs
func (s *Server) SyncLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get recent sync logs (last 20)
	logs, err := s.jobStore.GetRecentSyncLogs(20)
	if err != nil {
		response := models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to fetch sync logs: %v", err),
			Data:    []models.SyncLog{},
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    logs,
	}

	json.NewEncoder(w).Encode(response)
}

// PluginStatusHandler handles GET /plugins/status - Get plugin health status
func (s *Server) PluginStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Define plugins with their ports
	plugins := []map[string]interface{}{
		{"name": "Arbetsförmedlingen", "port": 8081, "id": "arbetsformedlingen"},
		{"name": "EURES", "port": 8082, "id": "eures"},
		{"name": "Remotive", "port": 8083, "id": "remotive"},
		{"name": "RemoteOK", "port": 8084, "id": "remoteok"},
		{"name": "Indeed Chrome", "port": 8087, "id": "indeed-chrome"},
		{"name": "Jooble", "port": 8088, "id": "jooble"},
	}

	// Get plugin URLs from environment
	pluginURLs := map[string]string{
		"arbetsformedlingen": os.Getenv("PLUGIN_ARBETSFORMEDLINGEN_URL"),
		"eures":              os.Getenv("PLUGIN_EURES_URL"),
		"remotive":           os.Getenv("PLUGIN_REMOTIVE_URL"),
		"remoteok":           os.Getenv("PLUGIN_REMOTEOK_URL"),
		"indeed-chrome":      os.Getenv("PLUGIN_INDEED_CHROME_URL"),
		"jooble":             os.Getenv("PLUGIN_JOOBLE_URL"),
	}

	// Check health of each plugin and get job count
	var pluginStatus []map[string]interface{}
	for _, plugin := range plugins {
		port := plugin["port"].(int)
		id := plugin["id"].(string)
		name := plugin["name"].(string)

		// Check health using environment URL or fallback to localhost
		baseURL := pluginURLs[id]
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%d", port)
		}
		healthURL := fmt.Sprintf("%s/health", baseURL)

		resp, err := http.Get(healthURL)
		status := "unhealthy"
		if err == nil && resp.StatusCode == 200 {
			status = "healthy"
			resp.Body.Close()
		}

		// Get job count for this connector from sync logs
		logs, _ := s.jobStore.GetRecentSyncLogs(100)
		jobCount := 0
		for _, log := range logs {
			if log.ConnectorName == id && log.Status == "success" {
				jobCount = log.JobsInserted
				break
			}
		}

		pluginStatus = append(pluginStatus, map[string]interface{}{
			"name":   name,
			"port":   port,
			"status": status,
			"jobs":   jobCount,
		})
	}

	response := models.APIResponse{
		Success: true,
		Data:    pluginStatus,
	}

	json.NewEncoder(w).Encode(response)
}
