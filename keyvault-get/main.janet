(use sh)
(import cmd)
(import spork/json)
(import spork/sh)

(defn read-config [config-file]
  (-> config-file
      (file/open :r)
      (file/read :all)))

(defn parse-config [config-file]
  (-> config-file
      (read-config)
      (json/decode)))

(defn write-config [config-file config]
  (-> config-file
      (file/open :w)
      (file/write config)
      (file/close)))

(defn create-config-file [config-file]
  (unless (sh/exists? config-file)
    (let [file-handle (sh/make-new-file config-file)]
      (file/write file-handle (json/encode @{}))
      (file/close file-handle))))

(defn list-secrets [keyvault]
  (json/decode ($< az keyvault secret list --vault-name ,keyvault --query "[].name")))

(defn fetch-secret [secret keyvault]
  (unless (= secret nil)
    ($ az keyvault secret show --name ,secret --vault-name ,keyvault --query value -o tsv |xclip -selection c)))

(defn get-secrets [config keyvault refresh]
  (let [cached-secrets (get config keyvault)]
    (if (or refresh (= cached-secrets nil))
      (list-secrets keyvault)
      cached-secrets)))

(defn select-secret [config-file keyvault refresh]
  (def config (parse-config config-file))
  (def secrets (get-secrets config keyvault refresh))
  (def secret @"")
  (set (config keyvault) secrets)
  # TODO: don't do this here
  (write-config config-file (json/encode config))
  (def [exit-status] (run fzf < ,(string/join secrets "\n") > ,secret))
  (case exit-status
    0 (string/trim (string secret))
    1 nil
    2 (error "fzf error")
    130 nil
    (error "unknown error")))

(cmd/main (cmd/fn
            "Fetch a secret from the Azure KeyVault"
            [--keyvault (required :string) "Name of the Azure KeyVault"
             --refresh (flag) "Refresh the cached list of secrets"
             secret (optional :string) "Secret to fetch"]
            (def config-file (string (os/getenv "HOME") "/" ".config/keyvault-get/config.json"))
            (create-config-file config-file)
            (fetch-secret (or secret (select-secret config-file keyvault refresh))
                          keyvault)))
