{
  description = "develop environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { nixpkgs, ... }:
    let
      # nixpkgs 26.11 and later no longer support x86_64-darwin.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
    in
    {
      devShells = nixpkgs.lib.genAttrs systems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          go = pkgs.go_1_27;
        in
        {
          default = pkgs.mkShell {
            packages = [ go ];

            # Print only `go: <version>` so `nix develop` stays terse.
            # Read the pinned package version instead of running `go env`,
            # which can switch toolchains (GOTOOLCHAIN=auto) or print nothing.
            shellHook = ''
              echo "go: ${go.version}"
            '';
          };
        }
      );
    };
}
