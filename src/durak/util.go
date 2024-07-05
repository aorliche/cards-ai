package durak

func IndexOf[T comparable](slice []T, val T) int {
    for i, v := range slice {
        if v == val {
            return i
        }
    }
    return -1
}
