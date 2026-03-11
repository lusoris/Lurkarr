# Questions & Answers Log

## 2026-03-11 — Initial Planning Round

### Logging
**Q:** Remove frontend log rendering?
**A:** Yes — remove the heavy log viewer from the frontend (like Huntarr/Cleanuparr did). Keep proper backend logging (slog, structured, state-of-the-art Go logging). Monitoring replaces frontend log rendering → use Grafana/Loki for log exploration instead.

**Action:** Remove `/logs` frontend page, keep slog + Loki integration. Evaluate if DB log storage + WebSocket broadcast are still needed for Loki or can be simplified.

### Notifications
**Q:** Remove notifications?
**A:** NO — notifications were ADDED and are a feature. They need to be reflected in the README and todo properly. 8 providers are implemented and working.

**Action:** Add notifications to README feature list. Frontend page for notification settings still needed (Phase 14).

### Coder Instance
**Q:** Coder URL?
**A:** `https://code.dev.cauda.dev`

**Q:** Provisioner/base?
**A:** Runs on Kubernetes. Template should target K8s pods (can also support Docker). Template determines the provisioner — create a proper K8s-based Coder template.

### Grafana Dashboards
**Q:** Scope?
**A:** Lurkarr metrics + Arr stack overview (Sonarr/Radarr/etc health/status panels).

### Uber FX
**Q:** Priority?
**A:** BEFORE feature work — refactor first, then build on clean DI foundation.

### Test Coverage
**Q:** Target?
**A:** 90%+ overall coverage.
