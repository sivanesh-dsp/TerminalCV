import { motion, useReducedMotion } from 'framer-motion';
import banner from '@/data/ascii/banner.txt?raw';
import bannerSmall from '@/data/ascii/banner-small.txt?raw';

/**
 * Responsive ASCII wordmark.
 * - Wide screens get the tall "ANSI Shadow" banner.
 * - Small screens get a compact banner that fits without horizontal scroll.
 */
export function AsciiLogo() {
  const reduced = useReducedMotion();
  return (
    <motion.div
      initial={reduced ? false : { opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
      aria-hidden="true"
      className="select-none text-term-accent"
    >
      <pre className="hidden overflow-hidden text-[10px] leading-[1.05] sm:block md:text-[13px] lg:text-base">
        {banner}
      </pre>
      <pre className="block overflow-hidden text-[10px] leading-[1.05] sm:hidden">
        {bannerSmall}
      </pre>
    </motion.div>
  );
}
