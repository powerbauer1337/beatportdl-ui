# BeatportDL Installation and Setup Guide (Windows 10)

This guide provides detailed instructions for installing and setting up BeatportDL on Windows 10. BeatportDL is a command-line tool that allows you to download music from Beatport in FLAC or AAC quality, provided you have a Beatport Streaming subscription.

## Prerequisites

Before installing BeatportDL, ensure you have the following:

*   **Beatport Streaming Subscription:** A valid Beatport Streaming subscription is required to download music.
*   **(Optional) Chocolatey Package Manager:** Chocolatey is a package manager for Windows that simplifies the installation of software. While not strictly required, it's highly recommended for installing dependencies.  If you don't have it, install it from [https://chocolatey.org/](https://chocolatey.org/). You'll need to run the installation command in an **administrator** PowerShell or Command Prompt.
*   **Git:** Git is used to clone the BeatportDL repository if you choose to build from source. Download and install it from [https://git-scm.com/download/win](https://git-scm.com/download/win).

## Installation

You can install BeatportDL using one of the following methods:

### Method 1: Download Pre-built Binaries (Recommended)

This is the easiest and fastest way to get started.

1.  **Download the Binary:**
    *   Go to the [BeatportDL Releases page](https://github.com/unspok3n/beatportdl/releases/).
    *   Download the appropriate binary for Windows (look for a file like `beatportdl-windows-amd64.exe`).

2.  **Place the Binary:**
    *   Choose a location to store the `beatportdl-windows-amd64.exe` file (e.g., `C:\Program Files\BeatportDL` or a folder in your user directory).

3.  **(Optional) Add to PATH:**
    *   To run BeatportDL from any location in the command line, add the directory containing the executable to your system's PATH environment variable.
    *   Search for "environment variables" in the Start menu and select "Edit the system environment variables."
    *   Click "Environment Variables..."
    *   In the "System variables" section (or "User variables" for a user-specific installation), find the "Path" variable and click "Edit...".
    *   Click "New" and add the full path to the directory where you placed `beatportdl-windows-amd64.exe` (e.g., `C:\Program Files\BeatportDL`).
    *   Click "OK" on all dialogs to save the changes.  You may need to restart your command prompt or PowerShell for the changes to take effect.

### Method 2: Build from Source

Building from source provides more flexibility but requires additional steps and dependencies.

1.  **Clone the Repository:**
    *   Open a command prompt or PowerShell window.
    *   Navigate to a directory where you want to store the BeatportDL source code.
    *   Clone the repository using Git:
```
bash
        git clone <repository_url>  # Replace with the actual repository URL
        cd beatportdl
        
```
2.  **Install Dependencies:**

    *   **Using Chocolatey (Recommended):** If you have Chocolatey installed, open an **administrator** PowerShell or Command Prompt and run:
```
bash
        choco install taglib zlib zig
        
```
*   **Manual Installation:** If you prefer manual installation:
        *   **TagLib:**
            *   Download the TagLib development package for Windows 64-bit from the official TagLib website or a trusted source.
            *   Extract the contents to a directory (e.g., `C:\Libraries\taglib`).
        *   **zlib:**
            *   Download the zlib development package for Windows 64-bit.
            *   Extract the contents to a directory (e.g., `C:\Libraries\zlib`).
        *   **Zig:**
            *   Download the Zig compiler for Windows (64-bit) from [https://ziglang.org/](https://ziglang.org/).
            *   Extract the contents to a directory (e.g., `C:\Zig`).
            *   Add the directory containing the `zig.exe` executable to your system's PATH environment variable (see instructions in "Add to PATH" in Method 1).

3.  **Configure Environment Variables:**

    *   Create a file named `.env` in the root directory of the cloned BeatportDL repository.
    *   Add the following line to the `.env` file, replacing the paths with the actual locations of the TagLib and zlib libraries and headers on your system.  Note that you may need to adjust these paths if you did not use the recommended install locations.
```
        WINDOWS_AMD64_LIB_PATH="-LC:/ProgramData/Chocolatey/lib/taglib/tools/lib -LC:/ProgramData/Chocolatey/lib/zlib/tools/lib -IC:/ProgramData/Chocolatey/lib/taglib/tools/include -IC:/ProgramData/Chocolatey/lib/zlib/tools/include"
        
```
**If you installed manually, adapt the paths accordingly:**
```
        WINDOWS_AMD64_LIB_PATH="-LC:/Libraries/taglib/lib -LC:/Libraries/zlib/lib -IC:/Libraries/taglib/include -IC:/Libraries/zlib/include"
        
```
**Important:** Use forward slashes (`/`) in the paths, even on Windows.

4.  **Build BeatportDL:**
    *   Open a command prompt or PowerShell window and navigate to the BeatportDL repository directory.
    *   Run the following command:
```
bash
        make windows-amd64
        
```
*   This will compile BeatportDL and place the executable (`beatportdl-windows-amd64.exe`) in the `bin/` directory.

## Setup and Configuration

1.  **Run BeatportDL:** Open a command prompt or PowerShell window and navigate to the directory containing the `beatportdl-windows-amd64.exe` file (either the location where you downloaded the binary or the `bin/` directory after building from source).  Run the executable:
```
bash
    ./beatportdl-windows-amd64.exe
    
```
**Note:** If you added the directory to your PATH, you can simply run:
```
bash
    beatportdl-windows-amd64.exe
    
```
2.  **Initial Configuration:** The first time you run BeatportDL, it will prompt you for:

    *   Beatport username
    *   Beatport password
    *   Downloads directory (where downloaded files will be saved)
    *   Audio quality (choose from: `medium-hls`, `medium`, `high`, `lossless`)

3.  **Configuration File (Optional):** BeatportDL will create a `beatportdl-config.yml` file in the same directory as the executable. You can customize various settings (e.g., download quality, file naming, tag mappings) by editing this file. Refer to the `README.md` file in the repository for a detailed description of the available options.

## Usage

1.  **Run BeatportDL:** Execute the `beatportdl-windows-amd64.exe` file.
2.  **Provide Input:**

    *   **Interactive Mode:** BeatportDL will prompt you to "Enter URL or search query:". You can enter:
        *   A Beatport URL (for a track, release, playlist, chart, label, or artist).
        *   A search query (BeatportDL will attempt to find matching results).

    *   **Command-line Arguments:** You can provide URLs directly as command-line arguments:
```
bash
        beatportdl-windows-amd64.exe <url1> <url2> ...
        
```
*   **Text File:**  You can provide a text file containing URLs (one URL per line) using the `-f` flag:
```
bash
        beatportdl-windows-amd64.exe -f urls.txt
        
```
3.  BeatportDL will then download the specified music according to your configuration and display progress information in the console.

## Troubleshooting

*   **"beatportdl-windows-amd64.exe' is not recognized..."**: If you get this error, it means the directory containing `beatportdl-windows-amd64.exe` is not in your PATH. Either navigate to that directory in the command prompt before running the command, or add the directory to your PATH as described in the Installation section.
*   **Build Errors (Method 2):** If you encounter errors during the build process, double-check the following:
    *   **Dependencies:** Ensure that TagLib, zlib, and Zig are correctly installed and that you have the development packages (including headers and libraries).
    *   **Environment Variables:** Verify that the `WINDOWS_AMD64_LIB_PATH` in your `.env` file points to the correct locations of the TagLib and zlib headers and libraries.  Use forward slashes in the paths.
    *   **Permissions:** If you installed dependencies manually and encounter permission errors during linking, ensure that the library directories and files have appropriate read permissions.
    *   **Zig Version:** Make sure you have a compatible version of Zig installed (check the `README.md` for any version requirements).
*   **DLL Errors:** If you get errors about missing DLL files (e.g., `tag.dll`, `zlib1.dll`) when running `beatportdl-windows-amd64.exe`, it means these libraries are not in a location where the executable can find them.  Try one of the following:
    *   **Copy DLLs:** Copy the missing DLL files from the TagLib and zlib library directories (e.g.,  `C:\ProgramData\Chocolatey\lib\taglib\tools\lib` and `C:\ProgramData\Chocolatey\lib\zlib\tools\lib` if you used Chocolatey, or your manual installation directories) to the same directory as `beatportdl-windows-amd64.exe` (the `bin/` directory if you built from source, or the directory where you downloaded the binary).
    *   **Add to PATH:** Add the directories containing the DLLs to your system's PATH environment variable.
*   **Network Issues:** If downloads fail, ensure you have a stable internet connection and that your firewall isn't blocking BeatportDL.  If you're using a proxy, configure it in the `beatportdl-config.yml` file.
*   **Beatport Subscription:** Double-check that your Beatport Streaming subscription is active.

If you continue to experience problems, consult the project's issue tracker or community forums for assistance. Please provide detailed information about the error messages you're seeing, your system configuration, and the steps you've taken.