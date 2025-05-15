{
  description = "A Nix-flake-based Go 1.24 development environment";

  inputs.nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1.*.tar.gz";

  outputs = { self, nixpkgs }:
    let
      goVersion = 24; # Change this to update the whole stack

      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ self.overlays.default ];
          config = {
            allowUnfree = true;
            # permittedInsecurePackages = [ "olm-3.2.16" ];
          };
        };
      });
    in
    {
      overlays.default = final: prev: {
        go = final."go_1_${toString goVersion}";
      };

      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {


          # If linking with libolm on macOS, then installing libolm via homebrew and setting
          # these two env vars is the only way I've gotten it to work. 
          # LIBRARY_PATH = "/opt/homebrew/lib";
          # CPATH = "/opt/homebrew/include";
          # If using olm via nix, then there seems to be no binary version of olm for darwin aarch64,
          # so it tries to build it from source. But the source is 9 years old and have constructs 
          # that does not compile with CC from nix on darwin. 
          # What I'm currently doing, is running "go [run|build] -tags goolm ...", which seems to work OK 
          # this far, even if the devs of mautrix don't recommend goolm for production yet.

          packages = with pkgs; [
            # go (version is specified by overlay)
            go

            # goimports, godoc, etc.
            gotools

            # https://github.com/golangci/golangci-lint
            golangci-lint
            golangci-lint-langserver
            gopls

            pkg-config

            # Needed for crypto
            # olm
          ];
        };
      });
    };
}
