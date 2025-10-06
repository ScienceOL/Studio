import clsx from 'clsx';

const NavbarFullWidthFooter = ({
  callsToAction,
  numberOfCallsToAction = 3,
}: {
  callsToAction: { name: string; href: string; icon: any }[];
  numberOfCallsToAction: number;
}) => {
  return (
    <div className=" bg-neutral-50/50 dark:bg-neutral-800/50">
      <div className="mx-auto max-w-7xl sm:px-6 lg:px-8">
        <div
          className={clsx(
            'sm:border-white/5` grid grid-cols-1 divide-y divide-neutral-900/5',
            'dark:divide-white/5 ',
            `sm:grid-cols-${numberOfCallsToAction} sm:divide-x sm:divide-y-0 sm:border-x sm:border-neutral-900/5`
          )}
        >
          {callsToAction.map((item) => (
            <a
              key={item.name}
              href={item.href}
              target="_blank"
              className="flex items-center gap-x-2.5 p-3 px-6
                 text-sm font-semibold leading-6 text-neutral-900 hover:bg-neutral-100
                  dark:text-white dark:hover:bg-neutral-700 sm:justify-center sm:px-0"
            >
              <item.icon
                className="h-5 w-5 flex-none text-neutral-400 dark:text-neutral-300"
                aria-hidden="true"
              />
              {item.name}
            </a>
          ))}
        </div>
      </div>
    </div>
  );
};

export default NavbarFullWidthFooter;
