// Jest setup provided by Grafana scaffolding
import './.config/jest-setup';

// jsdom has no canvas, which @grafana/ui uses to size the Combobox input
HTMLCanvasElement.prototype.getContext = () => ({
  measureText: (text) => ({ width: String(text).length * 8 }),
});
