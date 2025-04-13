# Check if MSYS2 is installed
if (-not (Test-Path "C:\msys64")) {
    Write-Host "MSYS2 is not installed in C:\msys64. Please run the installer first."
    exit 1
}

Write-Host "Updating MSYS2 packages..."
C:\msys64\usr\bin\bash.exe -lc 'pacman -Syu --noconfirm'

Write-Host "Installing MinGW-w64 GCC..."
C:\msys64\usr\bin\bash.exe -lc 'pacman -S --noconfirm mingw-w64-x86_64-gcc'

Write-Host "Installation complete. Please restart your terminal to use the new tools." 