{
  description = "Lilium CLI — ToolBear OIDC client (gh-style)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        lilium = pkgs.buildGoModule {
          pname = "lilium";
          version = "0.1.0";
          src = pkgs.lib.cleanSourceWith {
            src = ./.;
            filter = path: type:
              let base = baseNameOf path;
              in
              # Drop nix metadata and the locally-built binary at repo root.
              !(pkgs.lib.hasSuffix ".nix" path)
              && base != "flake.lock"
              && base != "result"
              && !(type == "regular" && base == "lilium")
              && base != ".git";
          };
          # Module path: github.com/lilium-toolbear/lilium-cli
          vendorHash = "sha256-7K17JaXFsjf163g5PXCb5ng2gYdotnZ2IDKk8KFjNj0=";
          subPackages = [ "cmd/lilium" ];
          ldflags = [ "-s" "-w" ];
          meta = with pkgs.lib; {
            description = "Lilium CLI for ToolBear OIDC-authenticated APIs";
            homepage = "https://github.com/lilium-toolbear/lilium-cli";
            license = licenses.mit;
            mainProgram = "lilium";
          };
        };
      in
      {
        packages.default = lilium;
        packages.lilium = lilium;

        apps.default = {
          type = "app";
          program = "${lilium}/bin/lilium";
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            golangci-lint
            git
          ];
        };
      });
}
