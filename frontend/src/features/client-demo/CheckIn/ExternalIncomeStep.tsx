import { useState, useEffect } from 'react';

export interface ExternalIncomeItemForm {
  amount: string;
  description: string;
}

interface ExternalIncomeStepProps {
  initialDeclaredNone?: boolean;
  initialItems?: ExternalIncomeItemForm[];
  onChange: (answered: boolean, none: boolean, items: ExternalIncomeItemForm[]) => void;
}

export function ExternalIncomeStep({
  initialDeclaredNone = false,
  initialItems = [],
  onChange,
}: ExternalIncomeStepProps) {
  const [selectedOption, setSelectedOption] = useState<'no' | 'yes' | null>(
    initialDeclaredNone ? 'no' : initialItems.length > 0 ? 'yes' : null,
  );
  const [items, setItems] = useState<ExternalIncomeItemForm[]>(
    initialItems.length > 0 ? initialItems : [{ amount: '', description: '' }],
  );

  useEffect(() => {
    if (selectedOption === 'no') {
      onChange(true, true, []);
    } else if (selectedOption === 'yes') {
      const valid = items.every(it => parseFloat(it.amount) > 0 && it.description.trim().length > 0);
      onChange(valid, false, items);
    } else {
      onChange(false, false, []);
    }
  }, [selectedOption, items, onChange]);

  const handleSelectOption = (opt: 'no' | 'yes') => {
    setSelectedOption(opt);
    if (opt === 'yes' && items.length === 0) {
      setItems([{ amount: '', description: '' }]);
    }
  };

  const handleItemChange = (index: number, field: keyof ExternalIncomeItemForm, val: string) => {
    const next = [...items];
    const current = items[index] ?? { amount: '', description: '' };
    next[index] = { ...current, [field]: val };
    setItems(next);
  };

  const handleAddItem = () => {
    setItems([...items, { amount: '', description: '' }]);
  };

  const handleRemoveItem = (index: number) => {
    if (items.length <= 1) {
      setItems([{ amount: '', description: '' }]);
      return;
    }
    setItems(items.filter((_, i) => i !== index));
  };

  return (
    <div className="rounded-2xl border border-neutral-200/80 bg-white p-4 shadow-sm space-y-3">
      <h3 className="text-sm font-bold text-neutral-900">
        ¿Cobraste algo fuera de Wallbit en octubre?
      </h3>

      <div className="space-y-2">
        <label
          className={`flex cursor-pointer items-center gap-3 rounded-xl border p-3 transition-colors ${
            selectedOption === 'no'
              ? 'border-neutral-900 bg-neutral-50/80'
              : 'border-neutral-200 hover:bg-neutral-50/50'
          }`}
        >
          <input
            type="radio"
            name="external-income"
            value="no"
            checked={selectedOption === 'no'}
            onChange={() => handleSelectOption('no')}
            className="h-4 w-4 accent-neutral-900"
          />
          <span className="text-xs font-medium text-neutral-800">
            No, todos mis cobros fueron a través de Wallbit
          </span>
        </label>

        <label
          className={`flex cursor-pointer items-center gap-3 rounded-xl border p-3 transition-colors ${
            selectedOption === 'yes'
              ? 'border-neutral-900 bg-neutral-50/80'
              : 'border-neutral-200 hover:bg-neutral-50/50'
          }`}
        >
          <input
            type="radio"
            name="external-income"
            value="yes"
            checked={selectedOption === 'yes'}
            onChange={() => handleSelectOption('yes')}
            className="h-4 w-4 accent-neutral-900"
          />
          <span className="text-xs font-medium text-neutral-800">
            Sí, cobré ingresos por fuera
          </span>
        </label>
      </div>

      {/* Dynamic item rows if "Sí" is chosen */}
      {selectedOption === 'yes' && (
        <div className="space-y-2.5 pt-2 border-t border-neutral-100">
          <p className="text-[11px] text-neutral-500">
            Ingresá el monto en US$ y una breve descripción de cada cobro:
          </p>
          {items.map((item, idx) => (
            <div key={idx} className="flex items-center gap-2">
              <div className="w-24">
                <input
                  type="text"
                  placeholder="US$ 0.00"
                  value={item.amount}
                  onChange={e => handleItemChange(idx, 'amount', e.target.value)}
                  className="w-full rounded-lg border border-neutral-300 px-2 py-1.5 text-xs focus:border-black focus:outline-none"
                  aria-label={`Monto en dólares cobro ${idx + 1}`}
                />
              </div>
              <div className="flex-1">
                <input
                  type="text"
                  placeholder="Descripción (ej. Asesoría directa)"
                  value={item.description}
                  onChange={e => handleItemChange(idx, 'description', e.target.value)}
                  className="w-full rounded-lg border border-neutral-300 px-2 py-1.5 text-xs focus:border-black focus:outline-none"
                  aria-label={`Descripción cobro ${idx + 1}`}
                />
              </div>
              {items.length > 1 && (
                <button
                  type="button"
                  onClick={() => handleRemoveItem(idx)}
                  className="rounded p-1 text-neutral-400 hover:text-red-600"
                  aria-label={`Eliminar cobro ${idx + 1}`}
                >
                  ✕
                </button>
              )}
            </div>
          ))}

          <button
            type="button"
            onClick={handleAddItem}
            className="text-xs font-medium text-neutral-800 underline underline-offset-2 hover:text-black"
          >
            + Agregar otro cobro
          </button>
        </div>
      )}
    </div>
  );
}
