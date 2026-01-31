package config

// UpdatePublicKey is the Minisign public key for verifying updates.
// This key corresponds to the private key used to sign official releases.
const UpdatePublicKey = "RWTx5Zr1YxbEJ+8b6LMk5z/6l4q1I9x8Jz3LqF9q2P6z1v8M1j5k3n7o" // Example/Placeholder key

// Version is the current application version.
// This should be updated during build via -ldflags or bumped manually.
var Version = "v0.1.0"
