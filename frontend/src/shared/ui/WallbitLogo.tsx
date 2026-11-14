import logoWhite from '../../assets/brand/wallbit-logo-white.svg';
import logoDark from '../../assets/brand/wallbit-logo-dark.svg';
import isotype from '../../assets/brand/wallbit-isotype.svg';

export function WallbitLogo({
  variant = 'white',
  className = 'h-6 w-auto',
}: {
  variant?: 'white' | 'dark' | 'isotype';
  className?: string;
}) {
  const src =
    variant === 'isotype'
      ? isotype
      : variant === 'dark'
        ? logoDark
        : logoWhite;

  return <img src={src} alt="Wallbit" className={className} />;
}
