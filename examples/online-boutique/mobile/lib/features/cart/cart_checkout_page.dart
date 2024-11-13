import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

final shippingNameProvider = StateProvider<String>((ref) => '');
final shippingAddressProvider = StateProvider<String>((ref) => '');
final shippingCityProvider = StateProvider<String>((ref) => '');
final shippingStateProvider = StateProvider<String>((ref) => '');
final shippingZipProvider = StateProvider<String>((ref) => '');
final creditCardNumberProvider = StateProvider<String>((ref) => '');
final creditCardExpirationProvider = StateProvider<String>((ref) => '');
final creditCardCvcProvider = StateProvider<String>((ref) => '');

class CartCheckoutPage extends HookConsumerWidget {
  final _formKey = GlobalKey<FormState>();
  CartCheckoutPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Checkout'),
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Shipping Information',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'Name',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your name';
                    }
                    return null;
                  },
                  onChanged: (value) =>
                      ref.read(shippingNameProvider.notifier).state = value,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'Address',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your address';
                    }
                    return null;
                  },
                  onChanged: (value) =>
                      ref.read(shippingAddressProvider.notifier).state = value,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'City',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your city';
                    }
                    return null;
                  },
                  onChanged: (value) =>
                      ref.read(shippingCityProvider.notifier).state = value,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'State',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your state';
                    }
                    return null;
                  },
                  onChanged: (value) =>
                      ref.read(shippingStateProvider.notifier).state = value,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'Zip',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your zip code';
                    }
                    return null;
                  },
                  onChanged: (value) =>
                      ref.read(shippingZipProvider.notifier).state = value,
                ),
                const SizedBox(height: 16.0),
                Text(
                  'Payment Information',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'Credit Card Number',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your credit card number';
                    }
                    return null;
                  },
                  onChanged: (value) =>
                      ref.read(creditCardNumberProvider.notifier).state = value,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'Expiration Date',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your credit card expiration date';
                    }
                    return null;
                  },
                  onChanged: (value) => ref
                      .read(creditCardExpirationProvider.notifier)
                      .state = value,
                ),
                const SizedBox(height: 16.0),
                TextFormField(
                  decoration: const InputDecoration(
                    labelText: 'CVC',
                  ),
                  validator: (value) {
                    if (value?.isEmpty == true) {
                      return 'Please enter your credit card CVC';
                    }
                    return null;
                  },
                  onChanged: (value) =>
                      ref.read(creditCardCvcProvider.notifier).state = value,
                ),
                const SizedBox(height: 16.0),
                Row(
                  children: [
                    ElevatedButton(
                      onPressed: () {
                        if (_formKey.currentState!.validate()) {}
                      },
                      child: const Text('Submit'),
                    ),
                    const Spacer(),
                    ElevatedButton(
                      onPressed: () {
                        if (_formKey.currentState!.validate()) {}
                      },
                      child: const Text('Prefill'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
