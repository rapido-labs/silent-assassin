package config

const KubernetesRunMode = "kubernetes.run_mode"

const ServerListenHost = "server_listen_host"
const ServerHost = "server_host"
const ServerPort = "server_port"
const Metrics = "/metrics"

const NodeSelectors = "label_selectors"
const ExpiryTimeAnnotation = "silent-assassin/expiry-time"

// const SpotterNodeSelectors = "spotter.label_selectors" TO BE REPLACED WITH NodeSelectors
const SpotterPollIntervalMs = "spotter.poll_interval_ms"

// const SpotterExpiryTimeAnnotation = "silent-assassin/expiry-time" TO BE REPLACED WITH ExpiryTimeAnnotation
const SpotterWhiteListIntervalHours = "spotter.white_list_interval_hours"

// const KillerDrainingTimeoutMs = "killer.draining_timeout_ms" TO BE REPLACED WITH KillerDrainingTimeoutWhenNodeExpiredMs
const KillerPollIntervalMs = "killer.poll_interval_ms"
const KillerDrainingTimeoutWhenNodeExpiredMs = "killer.draining_timeout_when_node_expired_ms"
const KillerDrainingTimeoutWhenNodePreemptedMs = "killer.draining_timeout_when_node_preempted_ms"

const ShifterPollIntervalMs = "shifter.poll_interval_ms"
const ShifterWhiteListIntervalHours = "shifter.white_list_interval_hours"
const ShifterNPResizeTimeout = "shifter.np_resize_timeout_mins"
const ShifterSleepAfterNodeDeletionMs = "shifter.sleep_after_node_deletion_ms"

const ClientServerRetries = "client.server_retries"
const ClientWatchMaintainanceEvents = "client.watch_maintainance_events"

const LogComponentName = "SILENT_ASSASSIN"
const LogLevel = "logger.level"

const SlackWebhookURL = "slack.webhook_url"
const SlackUsername = "slack.username"
const SlackChannel = "slack.channel"
const SlackIconURL = "slack.slack_icon_url"
const SlackTimeoutMs = "slack.slack_timeout"

const EventGetNodes = "GET_NODES"
const EventAnnotate = "ANNOTATE"
const EventDrain = "DRAIN"
const EventCordon = "CORDON"
const EventDeleteNode = "DELETE NODE"
const EventDeleteInstance = "DELETE INSTANCE"
const EventShift = "SHIFT"
const EventResizeNodePool = "RESIZE_NP"

const CommaSeparater = ","

const EvacuatePodsURI = "/evacuatepods"
