#ifndef SourceRoot
 #error SourceRoot required
#endif
[Setup]
AppId={{549647D8-C689-4C7F-A6A5-AB399CE61FEF}
AppName=PageSpindle
AppVersion=0.2.0
AppPublisher=Miguel Guhlin
AppPublisherURL=https://github.com/mguhlin/pagespindle
DefaultDirName={autopf}\PageSpindle
DefaultGroupName=PageSpindle
UninstallDisplayIcon={app}\pagespindle.exe
OutputDir={#SourceRoot}\dist
OutputBaseFilename=pagespindle-windows-amd64-setup
Compression=lzma2
SolidCompression=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
WizardStyle=modern
CloseApplications=yes
LicenseFile={#SourceRoot}\LICENSE
SetupLogging=yes
InfoBeforeFile={#SourceRoot}\packaging\windows\runtime-notice.txt
[Files]
Source: "{#SourceRoot}\dist\windows-amd64\PageSpindle\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "{#SourceRoot}\.cache\MicrosoftEdgeWebView2RuntimeInstallerX64.exe"; Flags: dontcopy
[Icons]
Name: "{group}\PageSpindle"; Filename: "{app}\pagespindle.exe"; WorkingDir: "{app}"
Name: "{autodesktop}\PageSpindle"; Filename: "{app}\pagespindle.exe"; WorkingDir: "{app}"
[Run]
Filename: "{app}\pagespindle.exe"; Description: "Open PageSpindle and download the local models (about 1.9 GB)"; Flags: nowait postinstall skipifsilent runasoriginaluser
[Code]
var NeedsRestart: Boolean;
function InstallRuntime(Name, Arguments: String): String;
var Code: Integer;
begin
 Result := '';
 if Name <> 'vc_redist.x64.exe' then ExtractTemporaryFile(Name);
 if not Exec(ExpandConstant('{tmp}\') + Name, Arguments, '', SW_HIDE, ewWaitUntilTerminated, Code) then
   Result := 'Cannot start ' + Name + '. Retry setup.'
 else if (Code = 3010) then NeedsRestart := True
 else if (Code <> 0) and (Code <> 1638) and (Code <> -2147023258) then
   Result := Name + ' could not finish (exit ' + IntToStr(Code) + '). Restart Windows and retry setup.';
end;
function PrepareToInstall(var NeedsRestartOut: Boolean): String;
begin
 try
  DownloadTemporaryFile('https://download.visualstudio.microsoft.com/download/pr/78d061c5-9ef3-42ab-98ed-3309d1095dc4/9C02C9ED120F671B900E6B9471D890E3B2414117A7E53F4AE2E855466FA2B7C9/VC_redist.x64.exe', 'vc_redist.x64.exe', '9c02c9ed120f671b900e6b9471d890e3b2414117a7e53f4ae2e855466fa2b7c9', nil);
 except
  Result := 'The Microsoft C++ runtime could not be downloaded and verified. Check the connection and retry setup.';
  Exit;
 end;
 Result := InstallRuntime('vc_redist.x64.exe', '/install /quiet /norestart');
 if (Result = '') and not RegKeyExists(HKLM32, 'SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}') and not RegKeyExists(HKCU, 'SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}') then
  Result := InstallRuntime('MicrosoftEdgeWebView2RuntimeInstallerX64.exe', '/silent /install');
 NeedsRestartOut := NeedsRestart;
end;
