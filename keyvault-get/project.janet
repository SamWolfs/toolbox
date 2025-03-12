(declare-project
  :name "keyvault-get"
  :description "Because AZ Keyvault kind of sucks"

  :dependencies ["https://github.com/janet-lang/spork.git" "https://github.com/ianthehenry/cmd.git"])

(declare-executable
  :name "keyvault-get"
  :entry "main.janet")
