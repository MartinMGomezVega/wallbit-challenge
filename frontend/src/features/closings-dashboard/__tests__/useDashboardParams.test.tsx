import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import React from 'react';
import {
  useDashboardParams,
  parseFiltersFromSearchParams,
  filtersToQueryString,
} from '../useDashboardParams';

function createWrapper(initialUrl = '/contador') {
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <MemoryRouter initialEntries={[initialUrl]}>
        {children}
      </MemoryRouter>
    );
  };
}

describe('useDashboardParams', () => {
  describe('pure parser and serializer', () => {
    it('parseFiltersFromSearchParams devuelve valores por defecto si no hay query', () => {
      const sp = new URLSearchParams('');
      const f = parseFiltersFromSearchParams(sp, '2026-10');
      expect(f.period).toBe('2026-10');
      expect(f.severity).toEqual([]);
      expect(f.regime).toBe('');
      expect(f.sort).toBe('severity');
      expect(f.dir).toBe('asc');
      expect(f.page).toBe(1);
      expect(f.pageSize).toBe(50);
    });

    it('parseFiltersFromSearchParams parsea parámetros complejos', () => {
      const sp = new URLSearchParams(
        'period=2026-09&severity=RED,YELLOW&regime=RI&q=Gomez&sort=toPay&dir=desc&page=2&pageSize=100'
      );
      const f = parseFiltersFromSearchParams(sp);
      expect(f.period).toBe('2026-09');
      expect(f.severity).toEqual(['RED', 'YELLOW']);
      expect(f.regime).toBe('RI');
      expect(f.q).toBe('Gomez');
      expect(f.sort).toBe('toPay');
      expect(f.dir).toBe('desc');
      expect(f.page).toBe(2);
      expect(f.pageSize).toBe(100);
    });

    it('filtersToQueryString serializa omitiendo valores por defecto', () => {
      expect(filtersToQueryString({})).toBe('');
      expect(
        filtersToQueryString({
          severity: ['RED'],
          page: 2,
        })
      ).toBe('?severity=RED&page=2');
    });
  });

  describe('hook in React Router', () => {
    it('lee filtros iniciales de la URL', () => {
      const wrapper = createWrapper('/contador?period=2026-10&severity=RED');
      const { result } = renderHook(() => useDashboardParams(), { wrapper });

      expect(result.current.filters.severity).toEqual(['RED']);
      expect(result.current.filters.period).toBe('2026-10');
    });

    it('toggleSeverity alterna colores y resetea página a 1', () => {
      const wrapper = createWrapper('/contador?page=3');
      const { result } = renderHook(() => useDashboardParams(), { wrapper });

      act(() => {
        result.current.toggleSeverity('RED');
      });

      expect(result.current.filters.severity).toEqual(['RED']);
      expect(result.current.filters.page).toBe(1);

      act(() => {
        result.current.toggleSeverity('YELLOW');
      });

      expect(result.current.filters.severity).toEqual(['RED', 'YELLOW']);

      act(() => {
        result.current.toggleSeverity('RED');
      });

      expect(result.current.filters.severity).toEqual(['YELLOW']);
    });

    it('setRegime, setQ y setSignal resetean la página a 1', () => {
      const wrapper = createWrapper('/contador?page=4');
      const { result } = renderHook(() => useDashboardParams(), { wrapper });

      act(() => {
        result.current.setRegime('MONOTRIBUTO');
      });
      expect(result.current.filters.regime).toBe('MONOTRIBUTO');
      expect(result.current.filters.page).toBe(1);

      act(() => {
        result.current.setPage(3);
      });
      expect(result.current.filters.page).toBe(3);

      act(() => {
        result.current.setQ('Lucia');
      });
      expect(result.current.filters.q).toBe('Lucia');
      expect(result.current.filters.page).toBe(1);
    });

    it('setSort invierte la dirección si se selecciona la misma columna', () => {
      const wrapper = createWrapper('/contador');
      const { result } = renderHook(() => useDashboardParams(), { wrapper });

      act(() => {
        result.current.setSort('billed');
      });
      expect(result.current.filters.sort).toBe('billed');
      expect(result.current.filters.dir).toBe('asc');

      act(() => {
        result.current.setSort('billed');
      });
      expect(result.current.filters.sort).toBe('billed');
      expect(result.current.filters.dir).toBe('desc');
    });

    it('clearFilters resetea todos los filtros de búsqueda preservando período', () => {
      const wrapper = createWrapper(
        '/contador?period=2026-08&severity=RED&q=Test&regime=RI&page=2'
      );
      const { result } = renderHook(() => useDashboardParams(), { wrapper });

      act(() => {
        result.current.clearFilters();
      });

      expect(result.current.filters.period).toBe('2026-08');
      expect(result.current.filters.severity).toEqual([]);
      expect(result.current.filters.q).toBe('');
      expect(result.current.filters.regime).toBe('');
      expect(result.current.filters.page).toBe(1);
    });
  });
});
