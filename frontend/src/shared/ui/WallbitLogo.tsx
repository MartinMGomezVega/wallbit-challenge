export function WallbitLogo({
  variant = 'white',
  className = 'h-6',
}: {
  variant?: 'white' | 'dark';
  className?: string;
}) {
  const textColor = variant === 'white' ? '#FFFFFF' : '#0A0A0A';
  return (
    <svg
      className={className}
      viewBox="0 0 110 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      role="img"
      aria-label="Wallbit"
    >
      {/* Isotipo azul de marca #0D99FF */}
      <path
        d="M2 3L7.5 19L12 8.5L16.5 19L22 3H18L14.5 14L11 5.5L7.5 14L4 3H2Z"
        fill="#0D99FF"
      />
      {/* Texto Wallbit */}
      <text
        x="27"
        y="17.5"
        fill={textColor}
        fontFamily="'Inter Variable', Inter, sans-serif"
        fontWeight="700"
        fontSize="16"
        letterSpacing="-0.3px"
      >
        wallbit
      </text>
    </svg>
  );
}
