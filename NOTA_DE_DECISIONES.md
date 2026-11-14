# Nota de decisiones

- **Qué hace la v1:** el día 1 arma sola el cierre de cada cliente (clasifica movimientos con reglas, concilia cobros con facturas, calcula Monotributo o RI y detecta el pase a RI) y lo marca 🔴/🟡/🟢. El contador revisa solo lo que no es verde, aprueba los verdes en lote y sigue el avance en un resumen; recién con su firma el cliente ve cuánto facturó, cuánto paga y hasta cuándo.
- **Qué dejé afuera:** login real, la vista del cliente (queda la API para la app de Wallbit), base gestionada en la nube con datos reales, emitir facturas y pagar en ARCA, integración real con ARCA y el tipo de cambio (mockeados con datos reales de 2026), Ingresos Brutos y DDJJ anuales.
- **Riesgo más grande:** que los clientes no carguen sus facturas a tiempo, porque es lo único que Wallbit no tiene y sin eso no hay cierre. Lo bajo con importación del CSV de ARCA y recordatorios; el paso siguiente es facturar desde Wallbit.
- **Cómo sabría si funcionó:** clientes por contador por mes (de ~30 a 300+), minutos de revisión por cliente, % de cierres verdes, % de clientes avisados antes del día 10 y % de cuotas pagadas en término; las tres últimas ya se ven en el resumen.
- **IA:** usé Claude para investigar la normativa 2026 y escribir las specs (SDD), y Gemini en Antigravity para implementarlas; las decisiones de producto y la revisión del código y los tests son míos.
