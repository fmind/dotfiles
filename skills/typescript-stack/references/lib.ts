export type Greeting = {
  readonly name: string;
};

export function greet({ name }: Greeting): string {
  return `Hello, ${name}!`;
}
