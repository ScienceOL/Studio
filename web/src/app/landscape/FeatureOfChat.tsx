import {
  BookOpenIcon,
  CloudArrowUpIcon,
  ServerIcon,
} from '@heroicons/react/20/solid';

const features = [
  {
    name: 'inclusiveness Discussions.',
    description:
      'Engage in open conversations with diverse members, fostering intellectual growth.',
    icon: CloudArrowUpIcon,
  },
  {
    name: 'Expert Insights.',
    description:
      'Access valuable knowledge from industry professionals and thought leaders.',
    icon: BookOpenIcon,
  },
  {
    name: 'Collaborative Learning.',
    description:
      "Tackle complex subjects, share resources, and support each other's quest for understanding.",
    icon: ServerIcon,
  },
];

export default function FeatureOfChat() {
  return (
    <div className="overflow-hidden bg-white py-24 dark:bg-neutral-950 sm:py-32">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <div className="mx-auto grid max-w-2xl grid-cols-1 gap-x-8 gap-y-16 sm:gap-y-20 lg:mx-0 lg:max-w-none lg:grid-cols-2">
          <div className="lg:pr-8 lg:pt-4">
            <div className="lg:max-w-lg">
              <h2 className="text-base font-semibold leading-7 text-indigo-600">
                Discussion to Research
              </h2>
              <p className="mt-2 text-3xl font-bold tracking-tight text-neutral-900 dark:text-neutral-50 sm:text-4xl">
                Join Freedom Space
              </p>
              <p className="mt-6 text-lg leading-8 text-neutral-600 dark:text-neutral-300">
                Embark on an intellectual journey with like-minded individuals
                in Freedom Space. Engage in stimulating discussions, exchange
                insights, and discover groundbreaking ideas in this vibrant
                community.
              </p>
              <dl className="mt-10 max-w-xl space-y-8 text-base leading-7 text-neutral-600 dark:text-neutral-300 lg:max-w-none">
                {features.map((feature) => (
                  <div key={feature.name} className="relative pl-9">
                    <dt className="inline font-semibold text-neutral-900 dark:text-neutral-50">
                      <feature.icon
                        className="absolute left-1 top-1 h-5 w-5 text-indigo-600"
                        aria-hidden="true"
                      />
                      {feature.name}
                    </dt>{' '}
                    <dd className="inline">{feature.description}</dd>
                  </div>
                ))}
              </dl>
            </div>
          </div>
          <img
            src="/Feature1.png"
            alt="Product screenshot"
            className="w-[48rem] max-w-none rounded-xl shadow-xl ring-1 ring-neutral-400/10 dark:ring-neutral-600/50 sm:w-[57rem] md:-ml-4 lg:-ml-0"
            width={2432}
            height={1442}
          />
        </div>
      </div>
    </div>
  );
}
