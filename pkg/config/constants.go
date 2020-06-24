package config

const KubernetesRunMode = "kubernetes.run_mode"

const SpotterNodeSelectors = "spotter.label_selectors"
const SpotterPollIntervalMs = "spotter.poll_interval_ms"
const SpotterExpiryTimeAnnotation = "silent-assassin/expiry-time"
const SpotterWhiteListIntervalHours = "spotter.white_list_interval_hours"

//Should be in ms or secs for uniformity??
const KillerDrainingTimeoutMs = "killer.draining_timeout_ms"
const KillerPollIntervalMs = "killer.poll_interval_ms"

const LogComponentName = "SILENT_ASSASSIN"
const LogLevel = "logger.level"
