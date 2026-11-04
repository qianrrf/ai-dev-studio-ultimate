; Electron's runtime registration has no uninstall hook. NSIS owns this product
; scheme and only removes a command that still points at this installation.
!macro customInstall
  WriteRegStr SHCTX "Software\Classes\ai-dev-studio-ultimate" "" "URL:AI Dev Studio Ultimate"
  WriteRegStr SHCTX "Software\Classes\ai-dev-studio-ultimate" "URL Protocol" ""
  WriteRegStr SHCTX "Software\Classes\ai-dev-studio-ultimate\DefaultIcon" "" '$INSTDIR\${APP_EXECUTABLE_FILENAME},0'
  WriteRegStr SHCTX "Software\Classes\ai-dev-studio-ultimate\shell\open\command" "" '"$INSTDIR\${APP_EXECUTABLE_FILENAME}" "%1"'
!macroend

!macro customUnInstall
  ReadRegStr $0 SHCTX "Software\Classes\ai-dev-studio-ultimate\shell\open\command" ""
  ${If} $0 == '"$INSTDIR\${APP_EXECUTABLE_FILENAME}" "%1"'
    DeleteRegKey SHCTX "Software\Classes\ai-dev-studio-ultimate"
  ${EndIf}
!macroend
