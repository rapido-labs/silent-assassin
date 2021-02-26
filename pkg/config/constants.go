package config

const KubernetesRunMode = "kubernetes.run_mode"

const ServerListenHost = "server_listen_host"
const ServerHost = "server_host"
const ServerPort = "server_port"
const Metrics = "/metrics"

const NodeSelectors = "label_selectors"
const ExpiryTimeAnnotation = "silent-assassin/expiry-time"

const SpotterPollInterval = "spotter.poll_interval"

const SpotterWhiteListIntervalHours = "spotter.white_list_interval_hours"

const KillerPollInterval = "killer.poll_interval"
const KillerDrainingTimeoutWhenNodeExpired = "killer.draining_timeout_when_node_expired"
const KillerDrainingTimeoutWhenNodePreempted = "killer.draining_timeout_when_node_preempted"
const KillerEvictDeleteDeadline = "killer.evict_delete_deadline"
const KillerGracePeriodSecondsWhenPodDeleted = "killer.grace_period_seconds_when_pod_deleted"

const ShifterEnabled = "shifter.enabled"
const ShifterPollInterval = "shifter.poll_interval"
const ShifterWhiteListIntervalHours = "shifter.white_list_interval_hours"
const ShifterNPResizeTimeout = "shifter.np_resize_timeout"
const ShifterSleepAfterNodeDeletion = "shifter.sleep_after_node_deletion"

const ClientServerRetries = "client.server_retries"
const ClientWatchMaintainanceEvents = "client.watch_maintainance_events"

const LogComponentName = "SILENT_ASSASSIN"
const LogLevel = "logger.level"

const SlackWebhookURL = "slack.webhook_url"
const SlackUsername = "slack.username"
const SlackChannel = "slack.channel"
const SlackIconURL = "slack.slack_icon_url"
const SlackTimeout = "slack.slack_timeout"

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
