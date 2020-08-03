package config

const KubernetesRunMode = "kubernetes.run_mode"

const ServerHost = "server_host"
const ServerPort = "server_port"

const SpotterNodeSelectors = "spotter.label_selectors"
const SpotterPollIntervalMs = "spotter.poll_interval_ms"
const SpotterExpiryTimeAnnotation = "silent-assassin/expiry-time"
const SpotterWhiteListIntervalHours = "spotter.white_list_interval_hours"

const KillerDrainingTimeoutMs = "killer.draining_timeout_ms"
const KillerPollIntervalMs = "killer.poll_interval_ms"
const KillerPreemptionDrainTimeout = "killer.preemption_timeout_ms"

const ClientServerRetries = "client.server_retries"
const ClientWatchMaintainanceEvents = "client.watch_maintainance_events"

const LogComponentName = "SILENT_ASSASSIN"
const LogLevel = "logger.level"

const SlackWebhookURL = "slack.webhook_url"
const SlackUsername = "slack.username"
const SlackChannel = "slack.channel"
const SlackIconURL = "slack.slack_icon_url"

const CommaSeparater = ","
