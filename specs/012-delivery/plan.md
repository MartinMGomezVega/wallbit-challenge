# 012 — Plan

## Borrador de la nota (ajustar con números reales del prototipo)
- **V1:** el día 1 arma solo el cierre de cada cliente (clasifica movimientos, concilia con facturas, calcula Monotributo/RI y detecta exclusión) y lo marca 🔴/🟡/🟢; el contador revisa solo lo que no es verde y aprueba en lote el resto; el cliente recibe el aviso recién después de la firma.
- **Afuera:** Ingresos Brutos, DDJJ anuales, emisión de facturas y pagos en ARCA (botones "Próximamente"), integración real con ARCA/BCRA, notas de crédito.
- **Riesgo:** que los clientes no carguen sus facturas a tiempo (sin facturas no hay cierre) y que una regla de clasificación mal calibrada genere ruido; mitigación: CSV de un click, recordatorios y reglas como parámetros del contador.
- **Éxito:** clientes por contador por mes (de 30 a 300+), % de cierres verdes, minutos de revisión por cliente, % de cierres notificados antes del día 10 y % de cuotas pagadas en término.
- **IA:** usé IA para investigar normativa 2026, escribir las specs (SDD) y generar código/tests a partir de ellas; las decisiones de producto y los datos los validé yo.

## Plan de 6 semanas (para el README)
| Semana | Foco |
|---|---|
| 1 | Datos reales de movimientos + reglas de clasificación con el contador; escalas/vencimientos reales. |
| 2 | Carga de facturas (CSV) + conciliación; piloto interno con 20 clientes. |
| 3 | Liquidación Monotributo + señales; tablero y modal de revisión. |
| 4 | RI + transición; aprobación auditada; avisos in-app/email. |
| 5 | Piloto con 100 clientes reales; calibrar umbrales; métricas. |
| 6 | Hardening (seguridad, backups, observabilidad) y lanzamiento a todos los que lo necesiten. |
