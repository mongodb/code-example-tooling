func validateInput(input string) error {
    if len(input) == 0 {
        return errors.New("input cannot be empty")
    }
    if len(input) > 100 {
        return errors.New("input too long")
    }
    return nil
}