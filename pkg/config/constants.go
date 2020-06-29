package config

const KubernetesRunMode = "kubernetes.run_mode"

const SpotterNodeSelectors = "spotter.label_selectors"
const SpotterPollIntervalMs = "spotter.poll_interval_ms"
const SpotterExpiryTimeAnnotation = "silent-assassin/expiry-time"

const KillerDeletionIntervalMs = "killer.deletion_interval"

//Should be in ms or secs for uniformity??
const KillerDrainingTimeoutSecs = "killer.draining_timeout"
const KillerPollIntervalMs = "killer.poll_interval_ms"

const LogComponentName = "SILENT_ASSASSIN"
const LogLevel = "logger.level"
