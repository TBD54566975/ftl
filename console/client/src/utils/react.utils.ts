type ClassNameProps = Array<string | undefined | false | null>;
/** takes an array of conditional css class name strings and returns them concatenated */
export const classNames = (...classes: ClassNameProps) =>
  classes.filter(Boolean).join(' ');
