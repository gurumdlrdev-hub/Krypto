Architecture Diagram
## System Architecture

```mermaid
flowchart TD
    Client["Client (Web Browser)"] -->|HTTP / Multipart Form Data| Router["Gin HTTP Router (:2409)"]
    Router --> Auth["PASETO Authentication Middleware"]

    %% Flow Split
    Auth -->|/encrypt/*| EncryptPipeline["Encrypt Engine"]
    Auth -->|/decrypt/*| DecryptPipeline["Decrypt Engine"]

    %% Encryption Flow
    subgraph Encryption ["Encryption Pipeline"]
        EncryptPipeline --> EncImgAud["Image / Audio Pipeline"]
        EncryptPipeline --> EncVid["Video Pipeline"]

        EncImgAud --> ReadRaw["Read Raw Binary Bytes"]
        ReadRaw --> AESEnc1["AES-256-GCM Encryption<br/>(Nonce Prepended + Auth Tag)"]
        AESEnc1 --> OutEnc[".enc Ciphertext Stream Download"]

        EncVid --> SplitVid["Frame Decomposition (30 FPS)<br/>+ Audio Stream Extraction"]
        SplitVid --> ReadFrames["Read Frame & Audio Bytes"]
        ReadFrames --> AESEnc2["AES-256-GCM Encryption"]
        AESEnc2 --> ZipArchive["Archive Encrypted Frames & Audio"]
        ZipArchive --> OutZip[".zip Archive Stream Download"]
    end

    %% Decryption Flow
    subgraph Decryption ["Decryption Pipeline"]
        DecryptPipeline --> DecImgAud["Image / Audio Pipeline"]
        DecryptPipeline --> DecVid["Video Pipeline"]

        DecImgAud --> ReadEnc["Parse .enc Stream<br/>(Extract 12-byte Nonce)"]
        ReadEnc --> AESDec1["AES-256-GCM Decryption<br/>(Tag Verification)"]
        AESDec1 --> OutMedia["Restored Media Stream<br/>(.png / .wav / .mp3)"]

        DecVid --> Unzip["Extract .zip Archive"]
        Unzip --> AESDec2["AES-256-GCM Decryption<br/>(Frames & Audio)"]
        AESDec2 --> StitchVid["Reassemble Frames (30 FPS)<br/>+ Merge Audio Track"]
        StitchVid --> OutMP4["Restored Video Stream (.mp4)"]
    end

    %% Egress
    OutEnc --> Client
    OutZip --> Client
    OutMedia --> Client
    OutMP4 --> Client
