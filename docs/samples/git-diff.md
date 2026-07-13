# Git operations

Given the following directory:

```
$ git status
On branch main

No commits yet

Changes to be committed:
  (use "git rm --cached <file>..." to unstage)
        new file:   file1.txt

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
        modified:   file1.txt

Untracked files:
  (use "git add <file>..." to include in what will be committed)
        new.txt
```

These are the outputs of useful commands in git:

```
$ git diff --numstat 
1       1       file1.txt

$ git diff --numstat --cached
2       0       file1.txt

$ git status --untracked-files --short
AM file1.txt
?? new.txt
```

I expect the diff tool to appear in the TUI as:

* git_diff('.') [+1/-1, +2/-0 staged, 1 new file]

Each part should only show up if they are there. For instance, if we only have
one new file and nothing else:

* git_diff('.') [1 new file]

If all changes are staged:

* git_diff('.') [+2/-0 staged]
