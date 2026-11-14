import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { useKeyboardNav } from '../useKeyboardNav';

function TestComponent(props: {
  enabled?: boolean;
  onNext?: () => void;
  onPrev?: () => void;
  onClose?: () => void;
  onApprove?: () => void;
}) {
  useKeyboardNav(props);
  return (
    <div>
      <input data-testid="test-input" />
      <textarea data-testid="test-textarea" />
      <div data-testid="editable" contentEditable />
    </div>
  );
}

describe('useKeyboardNav', () => {
  it('responde a flechas izquierda, derecha, escape y tecla A', () => {
    const onNext = vi.fn();
    const onPrev = vi.fn();
    const onClose = vi.fn();
    const onApprove = vi.fn();

    render(
      <TestComponent
        onNext={onNext}
        onPrev={onPrev}
        onClose={onClose}
        onApprove={onApprove}
      />
    );

    fireEvent.keyDown(window, { key: 'ArrowLeft' });
    expect(onPrev).toHaveBeenCalledTimes(1);

    fireEvent.keyDown(window, { key: 'ArrowRight' });
    expect(onNext).toHaveBeenCalledTimes(1);

    fireEvent.keyDown(window, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);

    fireEvent.keyDown(window, { key: 'a' });
    expect(onApprove).toHaveBeenCalledTimes(1);

    fireEvent.keyDown(window, { key: 'A' });
    expect(onApprove).toHaveBeenCalledTimes(2);
  });

  it('no dispara atajos si el foco está dentro de un input, textarea o contentEditable', () => {
    const onNext = vi.fn();
    const onPrev = vi.fn();
    const onClose = vi.fn();
    const onApprove = vi.fn();

    render(
      <TestComponent
        onNext={onNext}
        onPrev={onPrev}
        onClose={onClose}
        onApprove={onApprove}
      />
    );

    const input = screen.getByTestId('test-input');
    fireEvent.keyDown(input, { key: 'ArrowLeft' });
    fireEvent.keyDown(input, { key: 'ArrowRight' });
    fireEvent.keyDown(input, { key: 'Escape' });
    fireEvent.keyDown(input, { key: 'a' });

    expect(onPrev).not.toHaveBeenCalled();
    expect(onNext).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
    expect(onApprove).not.toHaveBeenCalled();

    const textarea = screen.getByTestId('test-textarea');
    fireEvent.keyDown(textarea, { key: 'a' });
    expect(onApprove).not.toHaveBeenCalled();

    const editable = screen.getByTestId('editable');
    fireEvent.keyDown(editable, { key: 'ArrowRight' });
    expect(onNext).not.toHaveBeenCalled();
  });

  it('no dispara nada si enabled es false', () => {
    const onNext = vi.fn();
    render(<TestComponent enabled={false} onNext={onNext} />);

    fireEvent.keyDown(window, { key: 'ArrowRight' });
    expect(onNext).not.toHaveBeenCalled();
  });

  it('no dispara onApprove si se presiona tecla A con modificadores (Ctrl, Meta, Alt)', () => {
    const onApprove = vi.fn();
    render(<TestComponent onApprove={onApprove} />);

    fireEvent.keyDown(window, { key: 'a', ctrlKey: true });
    fireEvent.keyDown(window, { key: 'a', metaKey: true });
    fireEvent.keyDown(window, { key: 'a', altKey: true });

    expect(onApprove).not.toHaveBeenCalled();
  });
});
