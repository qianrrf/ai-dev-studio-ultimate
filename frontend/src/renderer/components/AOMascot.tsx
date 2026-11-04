import aoMascot from "../../../assets/ultimate-mark.svg";

export function AOMascot({ className }: { className: string }) {
	return (
		<span className={`relative inline-block ${className}`} aria-hidden="true">
			<img className="relative size-full object-contain" src={aoMascot} alt="" />
		</span>
	);
}
