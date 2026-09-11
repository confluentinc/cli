package config

// Overrides the directory under $HOME that Flink statement history is written to. The default is
// no longer a constant here: it follows the build's release channel, via config.StateDirName.
const HomeConfluentPathEnvVar = "HOME_CONFLUENT_PATH"

// HomeConfluentPathDefault is the legacy stable-channel state directory name.
//
// Deprecated: the state directory now follows the build's release channel; use
// config.StateDirName instead. Retained as an exported alias for source compatibility.
const HomeConfluentPathDefault = ".confluent"
