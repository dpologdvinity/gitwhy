# Regenerate the README demo: vhs docs/demo.tape  (needs gitwhy on PATH)
Output docs/demo.gif
Set Shell bash
Set FontSize 16
Set Width 1280
Set Height 820
Set Padding 24
Set TypingSpeed 55ms
Set Theme "Catppuccin Mocha"

Hide
Type "rm -rf /tmp/gitwhy-demo && mkdir /tmp/gitwhy-demo && cd /tmp/gitwhy-demo && git init -q -b main && git config user.name Demo && git config user.email demo@example.com"
Enter
Type "echo '# app' > README.md && git add . && git commit -qm 'Initial commit' && mkdir src && echo 'login()' > src/auth.js && git add . && git commit -qm 'Add login'"
Enter
Type "echo 'logout()' >> src/auth.js && git commit -qam 'Add logout' && export PS1='\[\e[1;35m\]$ \[\e[0m\]' && clear"
Enter
Sleep 1s
Show

Type "gitwhy undo last commit"
Sleep 500ms
Enter
Sleep 2.5s
Enter
Sleep 2s
Down
Sleep 700ms
Up
Sleep 700ms
Enter
Sleep 5s
Enter
Sleep 3s
Type "git status --short"
Enter
Sleep 2s
Type "gitwhy undo"
Enter
Sleep 3s
Enter
Sleep 2.5s
Type "git log --oneline"
Enter
Sleep 3.5s
