let
  nixpkgs = import (
    let
      version = "66e66b9d481386f4c7a6b998f32ce794af0182ae";
    in builtins.fetchTarball {
      name   = "nixpkgs-${version}";
      url    = "https://github.com/NixOS/nixpkgs/archive/${version}.tar.gz";
      sha256 = "1zw3ib0705n7nskv9a2ipj1z0ys4wn2j8frnzhf1gx1yrgyjm8sn";
      }
    ) {};

in nixpkgs.mkShell {
  buildInputs = [
    nixpkgs.go
  ];
}
