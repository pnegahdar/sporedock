package utils


func FlattenHashes(hashes ...map[string]string) map[string]string {
    final := map[string]string{}
    for _, hash := range (hashes) {
        for k, v := range (hash) {
            final[k] = v
        }
    }
    return final
}

