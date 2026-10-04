# linter-autoconfigure-sdk — flake following the go-standard template
# (go-nix-helpers/templates/go-standard). Provides packages, apps,
# devShells (default/ci), and checks via the go-standard module.
#
# All larsartmann deps in go.mod are PUBLIC (served by proxy.golang.org,
# verified 2026-10-04), so they are declared via publicDeps instead of
# git+ssh inputs + deps — this keeps `nix build` working in CI without
# SSH keys and skips the mkPreparedSource machinery entirely.
{
  description = "Shared foundation for linter auto-configuration tools — config round-trip, finding emission, and a provider spec for BuildFlow integration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "linter-autoconfigure-sdk";
        vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # nix build to compute
        description = "Shared foundation for linter auto-configuration tools";

        # go.mod floor is `go 1.27`; the module default (go_1_26 = 1.26.7)
        # fails under the devShell's GOTOOLCHAIN=local.
        goPkgAttr = "go_1_27";

        # Match the private pattern but are actually public — excluded from
        # validatePrivateDeps; resolved via proxy.golang.org in the FOD.
        publicDeps = [
          "github.com/larsartmann/go-atomic-write"
          "github.com/larsartmann/go-finding"
          "github.com/larsartmann/go-error-family"
        ];

        # TestREADMESnippetsCompile compiles README snippets in a temp module
        # that cannot resolve private deps under the sandbox's GOPROXY=off
        # (locally they resolve via direct VCS fetch). Same class as pdd's
        # check disablement: tests run in the devShell, not the hermetic build.
        enableCheck = false;
      };
    };
}
