/**
 * React component name helper
 */

export const toComponentName = (name: string): string => {
  return String(name)
    .split(/[-_]/)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join('');
};
