package version

// Version is the semantic version of cortex (without a leading "v").
// It is updated by scripts/release.sh and can also be overridden at link time
// via -ldflags "-X github.com/mrvicadai/cortex/internal/version.Version=X.Y.Z".
var Version = "0.4.2"
