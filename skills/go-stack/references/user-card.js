// One Alpine component per module, named after its file. Components take the
// server-rendered Go struct as their argument — templ.JSFuncCall safely serializes
// it, so the shape here mirrors the Go type exactly.
export function userCard(user) {
  return {
    user,
    expanded: false,

    toggle() {
      this.expanded = !this.expanded;
    },
  };
}
